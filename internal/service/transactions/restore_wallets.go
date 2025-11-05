package transactions

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_stores"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_update_balance_queue"

	commonv2 "github.com/dv-net/dv-proto/gen/go/eproxy/common/v2"
	transactionsv2 "github.com/dv-net/dv-proto/gen/go/eproxy/transactions/v2"

	"connectrpc.com/connect"
	"github.com/dv-net/dv-processing/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type TxRestorer interface {
	RestoreWallet(ctx context.Context, ID uuid.UUID) error
	RestoreAllWallets(ctx context.Context, blockchains []string) error
}

func (s *Service) RestoreAllWallets(ctx context.Context, blockchains []string) error {
	res, err := s.storage.WalletAddresses().GetWalletsDataForRestoreByBlockchains(ctx, blockchains)
	if err != nil {
		return fmt.Errorf("get all wallet addresses: %w", err)
	}

	const maxWalletWorkers = 20

	sem := make(chan struct{}, maxWalletWorkers)
	wg := sync.WaitGroup{}
	for _, v := range res {
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer func() {
				wg.Done()
				<-sem
			}()

			if err := s.syncWalletTransactions(ctx, v.WalletAddress, v.Store, v.Currency); err != nil {
				s.log.Errorw("restore wallet transactions", "error", err)
			}
		}()
	}

	wg.Wait()
	return nil
}

func (s *Service) RestoreWallet(ctx context.Context, id uuid.UUID) error {
	wallet, err := s.storage.WalletAddresses().GetById(ctx, id)
	if err != nil {
		return fmt.Errorf("find wallet: %w", err)
	}

	store, err := s.storage.Stores().GetStoreByWalletAddress(ctx, repo_stores.GetStoreByWalletAddressParams{
		Address:    wallet.Address,
		CurrencyID: wallet.CurrencyID,
	})
	if err != nil {
		return fmt.Errorf("find store by wallet: %w", err)
	}

	curr, err := s.storage.Currencies().GetByID(ctx, wallet.CurrencyID)
	if err != nil {
		return fmt.Errorf("find curr: %w", err)
	}

	return s.syncWalletTransactions(ctx, *wallet, *store, *curr)
}

func (s *Service) syncWalletTransactions(ctx context.Context, wallet models.WalletAddress, store models.Store, curr models.Currency) error {
	if !curr.ContractAddress.Valid {
		return errors.New("currency has no contract address")
	}

	blockchainEPB, err := wallet.Blockchain.ToEPb()
	if err != nil {
		return fmt.Errorf("get epb: %w", err)
	}

	var currentPage uint32 = 1
	nextPageExists := true
	for nextPageExists {
		nextPageExists, err = s.fetchAndProcessTransactionsPage(ctx, wallet, curr, store, blockchainEPB, currentPage)
		if err != nil {
			return err
		}

		currentPage++
	}
	return nil
}

func (s *Service) fetchAndProcessTransactionsPage(ctx context.Context, wallet models.WalletAddress, curr models.Currency, store models.Store, blockchainEPB commonv2.Blockchain, page uint32) (bool, error) {
	const defaultPageSize uint32 = 50
	txs, err := s.epr.Transactions().Find(ctx, connect.NewRequest(&transactionsv2.FindRequest{
		Blockchain:      blockchainEPB,
		ContractAddress: utils.Pointer(curr.ContractAddress.String),
		Common: &commonv2.FindRequestCommon{
			Page:     utils.Pointer(page),
			PageSize: utils.Pointer(defaultPageSize),
		},
		Address: utils.Pointer(wallet.Address),
	}))
	if err != nil {
		return false, fmt.Errorf("eproxy: %w", err)
	}

	for _, tx := range txs.Msg.GetItems() {
		if err := s.processSingleTransaction(ctx, wallet, curr, store, tx); err != nil {
			return txs.Msg.GetNextPageExists(), err
		}
	}

	return txs.Msg.GetNextPageExists(), nil
}

func (s *Service) processSingleTransaction(ctx context.Context, wallet models.WalletAddress, curr models.Currency, store models.Store, tx *transactionsv2.Transaction) error {
	txFee, err := decimal.NewFromString(tx.GetFee())
	if err != nil {
		return fmt.Errorf("convert fee to decimal: %w", err)
	}

	for _, ev := range tx.Events {
		err := s.processEvent(ctx, wallet, curr, store, tx, ev, txFee)
		if errors.Is(err, ErrTransactionAlreadyExists) {
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) processEvent(ctx context.Context, wallet models.WalletAddress, curr models.Currency, store models.Store, tx *transactionsv2.Transaction, ev *transactionsv2.Event, txFee decimal.Decimal) error {
	// Skip irrelevant events
	if ev.Type == nil || *ev.Type != transactionsv2.EventType_EVENT_TYPE_TRANSFER {
		return nil
	}

	// Skip irrelevant assets
	if ev.AssetIdentifier == nil || *ev.AssetIdentifier != curr.ContractAddress.String {
		return nil
	}

	// Skip zero transfers from other addresses
	if ev.GetValue() == "0" && ev.GetAddressFrom() != "" && ev.GetAddressFrom() != tx.GetAddressFrom() {
		return nil
	}

	addrFrom, addrTo := ev.GetAddressFrom(), ev.GetAddressTo()
	if addrTo != wallet.Address && addrFrom != wallet.Address {
		return nil // Skip transactions not involving the wallet
	}

	var txType models.TransactionsType
	if addrTo == wallet.Address {
		txType = models.TransactionsTypeDeposit
	} else if addrFrom == wallet.Address {
		txType = models.TransactionsTypeTransfer
	}

	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(dbTx pgx.Tx) error {
		_, err := s.GetByHashAndBcUniq(ctx, tx.Hash, ev.GetBlockchainUniqKey(), repos.WithTx(dbTx))
		if err == nil {
			return ErrTransactionAlreadyExists
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("find tx: %w", err)
		}

		amount, err := decimal.NewFromString(ev.GetValue())
		if err != nil {
			return fmt.Errorf("prepareTxAmount: %w", err)
		}

		amntUsd, err := s.conv.Convert(ctx, currconv.ConvertDTO{
			Source: store.RateSource.String(),
			From:   curr.Code,
			To:     models.CurrencyCodeUSD,
			Amount: amount.String(),
		})
		if err != nil {
			return fmt.Errorf("prepare tx amount usd: %w", err)
		}

		_, err = s.CreateTransaction(ctx, repo_transactions.CreateParams{
			UserID:             wallet.UserID,
			StoreID:            uuid.NullUUID{UUID: store.ID, Valid: true},
			AccountID:          uuid.NullUUID{UUID: wallet.AccountID.UUID, Valid: true},
			CurrencyID:         wallet.CurrencyID,
			Blockchain:         wallet.Blockchain.String(),
			TxHash:             tx.Hash,
			BcUniqKey:          ev.BlockchainUniqKey,
			Type:               txType,
			FromAddress:        addrFrom,
			ToAddress:          addrTo,
			Amount:             amount,
			AmountUsd:          decimal.NullDecimal{Decimal: amntUsd, Valid: true},
			Fee:                txFee,
			WithdrawalIsManual: false,
			NetworkCreatedAt:   pgtype.Timestamp{Time: tx.CreatedAt.AsTime(), Valid: true},
		}, repos.WithTx(dbTx))
		if err != nil {
			return fmt.Errorf("create transaction: %w", err)
		}

		addrForUpdateBalance := addrTo
		nativeCurrency, err := curr.Blockchain.NativeCurrency()
		if err != nil {
			return fmt.Errorf("get native currency: %w", err)
		}

		updateNativeTokenRequired := nativeCurrency == curr.ID && curr.Blockchain.RecalculateNativeBalance()
		if txType == models.TransactionsTypeTransfer {
			addrForUpdateBalance = addrFrom
			updateNativeTokenRequired = txFee.IsPositive() && curr.Blockchain.RecalculateNativeBalance()
		}

		_, err = s.storage.UpdateBalanceQueue(repos.WithTx(dbTx)).Create(ctx, repo_update_balance_queue.CreateParams{
			CurrencyID:               curr.ID,
			Address:                  addrForUpdateBalance,
			NativeTokenBalanceUpdate: updateNativeTokenRequired,
		})
		if err != nil {
			return fmt.Errorf("enqueue address balance: %w", err)
		}
		return nil
	})
}
