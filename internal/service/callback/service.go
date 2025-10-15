package callback

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/processing_request"
	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/receipts"
	"github.com/dv-net/dv-merchant/internal/service/store"
	"github.com/dv-net/dv-merchant/internal/service/transactions"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_receipts"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transfer_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transfers"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_unconfirmed_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_update_balance_queue"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type ICallback interface {
	HandleDepositCallback(ctx context.Context, dto DepositWebhookDto) error
	HandleTransferCallback(ctx context.Context, dto TransferWebhookDto) error
	HandleUpdateTransferStatusCallback(ctx context.Context, dto processing_request.TransferStatusWebhook) error
}

type Service struct {
	log                            logger.Logger
	eventListener                  event.IListener
	storage                        storage.IStorage
	transactionsService            transactions.ITransaction
	unconfirmedTransactionsService transactions.IUnconfirmedTransaction
	storeService                   store.IStore
	currConvService                currconv.ICurrencyConvertor
	receiptsService                receipts.IReceiptService
}

func New(
	logger logger.Logger,
	eventListener event.IListener,
	storage storage.IStorage,
	transactionsService transactions.ITransaction,
	unconfirmedTransactionsService transactions.IUnconfirmedTransaction,
	storeService store.IStore,
	currConvService currconv.ICurrencyConvertor,
	receiptsService receipts.IReceiptService,
) ICallback {
	return &Service{
		log:                            logger,
		eventListener:                  eventListener,
		storage:                        storage,
		transactionsService:            transactionsService,
		unconfirmedTransactionsService: unconfirmedTransactionsService,
		storeService:                   storeService,
		currConvService:                currConvService,
		receiptsService:                receiptsService,
	}
}

func (s *Service) HandleDepositCallback(ctx context.Context, dto DepositWebhookDto) error { //nolint:funlen
	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		amount, err := decimal.NewFromString(dto.Amount)
		if err != nil {
			return fmt.Errorf("parse amount: %w", err)
		}
		fee, err := decimal.NewFromString(dto.Fee)
		if err != nil {
			return fmt.Errorf("parse fee: %w", err)
		}

		if amount.IsZero() {
			s.log.Infof("amount is zero tx_hash: %s", dto.Amount)
			return nil
		}

		_, err = s.transactionsService.GetByHashAndBcUniq(ctx, dto.Hash, dto.TxUniqKey, repos.WithTx(tx))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("transaction found error: %w", err)
		}

		if err == nil {
			s.log.Info("transaction already exists")
			return nil
		}

		if dto.ExternalWalletID == nil && dto.WalletType == models.WalletTypeHot {
			return fmt.Errorf("invalid wallet id")
		}

		// todo add wallet to dto
		wallet, err := s.storage.Wallets(repos.WithTx(tx)).GetById(ctx, *dto.ExternalWalletID)
		if err != nil {
			return fmt.Errorf("wallet found error: %w", err)
		}

		storeData, err := s.storeService.GetStoreByWalletAddress(ctx, dto.ToAddress, dto.Currency.ID, repos.WithTx(tx))
		if err != nil {
			return fmt.Errorf("storeData found error: %w", err)
		}

		usdAmount, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
			Source:     storeData.RateSource.String(),
			From:       dto.Currency.Code,
			To:         models.CurrencyCodeUSDT,
			Amount:     amount.String(),
			StableCoin: dto.Currency.IsStablecoin,
			Scale:      &storeData.RateScale,
		})
		if err != nil {
			return fmt.Errorf("convert usd: %w", err)
		}

		exchangeRate := decimal.NewFromFloat(usdAmount.InexactFloat64() / amount.InexactFloat64()).Round(int32(dto.Currency.Precision))

		usdFee := decimal.Zero
		if !fee.IsZero() {
			usdFee, err = s.currConvService.Convert(ctx, currconv.ConvertDTO{
				Source:     storeData.RateSource.String(),
				From:       dto.Currency.Code,
				To:         models.CurrencyCodeUSDT,
				Amount:     fee.String(),
				StableCoin: dto.Currency.IsStablecoin,
				Scale:      &storeData.RateScale,
			})
			if err != nil {
				return fmt.Errorf("convert usd fee: %w", err)
			}
		}

		// Either take untrusted email or trusted email. Trusted email has priority.
		var userEmail string
		if wallet.UntrustedEmail.Valid {
			userEmail = wallet.UntrustedEmail.String
		}
		if wallet.Email.Valid {
			userEmail = wallet.Email.String
		}

		if s.checkUnconfirmed(dto) {
			uTransaction, err := s.createUnconfirmedTransaction(ctx, dto, storeData, wallet, amount, usdAmount, repos.WithTx(tx))
			if err != nil {
				return fmt.Errorf("unconfirmed transaction creation: %w", err)
			}

			if dto.IsSystem {
				return nil
			}

			err = s.eventListener.Fire(transactions.DepositUnconfirmedEvent{
				Tx:              *uTransaction,
				Store:           *storeData,
				Currency:        *dto.Currency,
				StoreExternalID: wallet.StoreExternalID,
				WebhookEvent:    models.WebhookEventPaymentNotConfirmed,
				DBTx:            tx,
			})
			if err != nil {
				s.log.Errorw("eventListener fire error", "error", err)
				return fmt.Errorf("eventListener fire error: %w", err)
			}
			return nil
		}
		var receiptID uuid.NullUUID

		if !dto.IsSystem {
			receipt, err := s.receiptsService.Create(ctx, repo_receipts.CreateParams{
				Status:     models.ReceiptStatusPaid,
				StoreID:    storeData.ID,
				CurrencyID: dto.Currency.ID,
				Amount:     amount,
				WalletID:   uuid.NullUUID{UUID: wallet.ID, Valid: true},
			}, repos.WithTx(tx))
			if err != nil {
				return fmt.Errorf("receipt creation: %w", err)
			}
			receiptID = uuid.NullUUID{UUID: receipt.ID, Valid: true}
		}

		createPrams := repo_transactions.CreateParams{
			UserID:             storeData.UserID,
			StoreID:            uuid.NullUUID{UUID: storeData.ID, Valid: true},
			ReceiptID:          receiptID,
			WalletID:           uuid.NullUUID{UUID: wallet.ID, Valid: true},
			CurrencyID:         dto.Currency.ID,
			Blockchain:         dto.Currency.Blockchain.String(),
			IsSystem:           dto.IsSystem,
			TxHash:             dto.Hash,
			BcUniqKey:          &dto.TxUniqKey,
			Type:               models.TransactionsTypeDeposit,
			FromAddress:        dto.FromAddress,
			ToAddress:          dto.ToAddress,
			Amount:             amount,
			AmountUsd:          decimal.NullDecimal{Decimal: usdAmount, Valid: true},
			Fee:                fee,
			WithdrawalIsManual: false,
			NetworkCreatedAt:   pgtype.Timestamp{Time: dto.NetworkCreatedAt, Valid: true},
		}

		transaction, err := s.transactionsService.CreateTransaction(ctx, createPrams, repos.WithTx(tx))
		if err != nil {
			return fmt.Errorf("transaction creation: %w", err)
		}

		nativeCurrency, err := dto.Currency.Blockchain.NativeCurrency()
		if err != nil {
			return fmt.Errorf("get native currency: %w", err)
		}

		nativeTokenUpdateRequired := nativeCurrency == dto.Currency.ID && dto.Blockchain.RecalculateNativeBalance()
		if err = s.enqueueAddressBalanceUpdate(ctx, dto, nativeTokenUpdateRequired, repos.WithTx(tx)); err != nil {
			return fmt.Errorf("cant't enqueue update balance: %w", err)
		}

		// if system transaction no send webhook
		if dto.IsSystem {
			return nil
		}
		// event for webhook
		err = s.eventListener.Fire(transactions.DepositReceivedEvent{
			Tx:              *transaction,
			Store:           *storeData,
			Currency:        *dto.Currency,
			StoreExternalID: wallet.StoreExternalID,
			WebhookEvent:    models.WebhookEventPaymentReceived,
			DBTx:            tx,
		})

		if !dto.IsSystem {
			err = s.eventListener.Fire(transactions.DepositReceiptSentEvent{
				Tx:              *transaction,
				Store:           *storeData,
				Currency:        *dto.Currency,
				StoreExternalID: wallet.StoreExternalID,
				DBTx:            tx,
				WalletEmail:     userEmail,
				WalletLocale:    wallet.Locale,
				ExchangeRate:    exchangeRate,
				UsdFee:          usdFee,
			})
		}
		if err != nil {
			s.log.Errorw("eventListener fire error", "error", err)
			return fmt.Errorf("eventListener fire error: %w", err)
		}

		return nil
	})
}

func (s *Service) HandleTransferCallback(ctx context.Context, dto TransferWebhookDto) error {
	// skip unconfirmed transaction
	if dto.Status == models.TransactionStatusWaitingConfirmations || dto.Status == models.TransactionStatusInMempool {
		return nil
	}

	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{IsoLevel: pgx.Serializable}, func(tx pgx.Tx) error {
		amount, err := decimal.NewFromString(dto.Amount)
		if err != nil {
			return fmt.Errorf("parse amount: %w", err)
		}

		if amount.IsZero() {
			s.log.Infof("amount is zero tx_hash: %s", dto.Amount)
			return nil
		}

		_, err = s.transactionsService.GetByHashAndBcUniq(ctx, dto.Hash, dto.TxUniqKey, repos.WithTx(tx))
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("transaction found error: %w", err)
		}

		if err == nil {
			s.log.Info("transaction already exists")
			return nil
		}

		if dto.TransferID.Valid {
			if err = s.storage.Transfers(repos.WithTx(tx)).UpdateTxHash(ctx, repo_transfers.UpdateTxHashParams{
				ID:     dto.TransferID.UUID,
				TxHash: util.Pointer(dto.Hash),
			}); err != nil {
				return fmt.Errorf("update transfer tx hash: %w", err)
			}
		}

		switch dto.WalletType {
		case models.WalletTypeProcessing:
			if !dto.TransferID.Valid {
				s.log.Infow(
					"withdrawal from processing transaction received without request id",
					"from",
					dto.FromAddress,
					"to",
					dto.ToAddress,
					"hash",
					dto.Hash,
				)

				return nil
			}

			withdrawalData, err := s.storage.WithdrawalsFromProcessing(repos.WithTx(tx)).FindByTransferID(
				ctx,
				dto.TransferID.UUID,
			)
			if err != nil {
				return fmt.Errorf("withdrawal from processing not found")
			}

			storeData, err := s.storeService.GetStoreByID(ctx, withdrawalData.WithdrawalFromProcessingWallet.StoreID)
			if err != nil {
				return fmt.Errorf("fetch store data: %w", err)
			}

			createParams, prepareTxErr := s.prepareCreateTxParamsByProcessingWallet(
				ctx,
				dto,
				storeData,
				amount,
			)
			if prepareTxErr != nil {
				return prepareTxErr
			}

			createdTx, err := s.transactionsService.CreateTransaction(ctx, *createParams, repos.WithTx(tx))
			if err != nil {
				return fmt.Errorf("transaction creation: %w", err)
			}
			return s.eventListener.Fire(transactions.WithdrawalFromProcessingReceivedEvent{
				WithdrawalID: withdrawalData.WithdrawalFromProcessingWallet.ID.String(),
				Tx:           *createdTx,
				Store:        *storeData,
				Currency:     *dto.Currency,
				WebhookEvent: models.WebhookEventWithdrawalFromProcessingReceived,
				DBTx:         tx,
			})
		case models.WalletTypeHot:
			createParams, prepareTxErr := s.prepareCreateTxParamsByHotWallet(ctx, dto, amount, tx)
			if prepareTxErr != nil {
				return prepareTxErr
			}

			_, err = s.transactionsService.CreateTransaction(ctx, *createParams, repos.WithTx(tx))
			if err != nil {
				return fmt.Errorf("transaction creation: %w", err)
			}

			// update balance native token if need
			updateNativeTokenRequired := createParams.Fee.IsPositive() && dto.Blockchain.RecalculateNativeBalance()
			if err = s.enqueueAddressBalanceUpdate(ctx, dto, updateNativeTokenRequired, repos.WithTx(tx)); err != nil {
				return fmt.Errorf("cant't update balance: %w", err)
			}
		default:
			return fmt.Errorf("unsupported wallet type: %s", dto.WalletType)
		}
		return nil
	})
}

func (s *Service) prepareCreateTxParamsByProcessingWallet(
	ctx context.Context,
	dto TransferWebhookDto,
	storeData *models.Store,
	amount decimal.Decimal,
) (*repo_transactions.CreateParams, error) {
	usdAmount, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
		Source:     storeData.RateSource.String(),
		From:       dto.Currency.Code,
		To:         "USDT",
		Amount:     amount.String(),
		StableCoin: dto.Currency.IsStablecoin,
		Scale:      &storeData.RateScale,
	})
	if err != nil {
		return nil, fmt.Errorf("convert usd: %w", err)
	}

	fee, err := decimal.NewFromString(dto.Fee)
	if err != nil {
		return nil, fmt.Errorf("parse fee: %w", err)
	}

	return &repo_transactions.CreateParams{
		UserID:             storeData.UserID,
		StoreID:            uuid.NullUUID{UUID: storeData.ID, Valid: true},
		CurrencyID:         dto.Currency.ID,
		Blockchain:         dto.Currency.Blockchain.String(),
		TxHash:             dto.Hash,
		BcUniqKey:          &dto.TxUniqKey,
		Type:               models.TransactionsTypeWithdrawalFromProcessing,
		FromAddress:        dto.FromAddress,
		ToAddress:          dto.ToAddress,
		Amount:             amount,
		AmountUsd:          decimal.NullDecimal{Decimal: usdAmount, Valid: true},
		Fee:                fee,
		WithdrawalIsManual: false,
		NetworkCreatedAt:   pgtype.Timestamp{Time: dto.NetworkCreatedAt, Valid: true},
	}, nil
}

func (s *Service) prepareCreateTxParamsByHotWallet(
	ctx context.Context,
	dto TransferWebhookDto,
	amount decimal.Decimal,
	tx pgx.Tx,
) (*repo_transactions.CreateParams, error) {
	if dto.ExternalWalletID == nil {
		return nil, fmt.Errorf("invalid external wallet id")
	}

	wallet, err := s.storage.Wallets(repos.WithTx(tx)).GetById(ctx, *dto.ExternalWalletID)
	if err != nil {
		return nil, fmt.Errorf("wallet found error: %w", err)
	}

	storeData, err := s.storeService.GetStoreByWalletAddress(ctx, dto.FromAddress, dto.Currency.ID, repos.WithTx(tx))
	if err != nil {
		return nil, fmt.Errorf("storeData found error: %w", err)
	}

	usdAmount, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
		Source:     storeData.RateSource.String(),
		From:       dto.Currency.Code,
		To:         models.CurrencyCodeUSDT,
		Amount:     amount.String(),
		StableCoin: dto.Currency.IsStablecoin,
		Scale:      &storeData.RateScale,
	})
	if err != nil {
		return nil, fmt.Errorf("convert usd: %w", err)
	}

	fee, err := decimal.NewFromString(dto.Fee)
	if err != nil {
		return nil, fmt.Errorf("parse fee: %w", err)
	}

	return &repo_transactions.CreateParams{
		UserID:             storeData.UserID,
		StoreID:            uuid.NullUUID{UUID: storeData.ID, Valid: true},
		WalletID:           uuid.NullUUID{UUID: wallet.ID, Valid: true},
		CurrencyID:         dto.Currency.ID,
		Blockchain:         dto.Currency.Blockchain.String(),
		TxHash:             dto.Hash,
		BcUniqKey:          &dto.TxUniqKey,
		Type:               models.TransactionsTypeTransfer,
		FromAddress:        dto.FromAddress,
		ToAddress:          dto.ToAddress,
		Amount:             amount,
		AmountUsd:          decimal.NullDecimal{Decimal: usdAmount, Valid: true},
		Fee:                fee,
		WithdrawalIsManual: false,
		NetworkCreatedAt:   pgtype.Timestamp{Time: dto.NetworkCreatedAt, Valid: true},
	}, nil
}

func (s *Service) HandleUpdateTransferStatusCallback(ctx context.Context, dto processing_request.TransferStatusWebhook) error {
	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.storage.Transfers(repos.WithTx(tx)).UpdateTransferStatus(ctx, repo_transfers.UpdateTransferStatusParams{
			Status:  dto.Status,
			Stage:   models.ResolveTransferStageByStatus(dto.Status),
			Step:    util.Pointer(dto.Step),
			ID:      *dto.RequestID,
			Message: dto.ErrorMessage,
		}); err != nil {
			return fmt.Errorf("update status callback: %w", err)
		}

		if len(dto.SystemTransactions) < 1 {
			return nil
		}

		batchParams := make([]repo_transfer_transactions.BatchCreateParams, 0, len(dto.SystemTransactions))
		for _, sysTx := range dto.SystemTransactions {
			batchParams = append(batchParams, repo_transfer_transactions.BatchCreateParams{
				TransferID:        *dto.RequestID,
				TxHash:            sysTx.TxHash,
				BandwidthAmount:   sysTx.BandwidthAmount,
				EnergyAmount:      sysTx.EnergyAmount,
				NativeTokenAmount: sysTx.NativeTokenAmount,
				NativeTokenFee:    sysTx.NativeTokenFee,
				TxType:            sysTx.TxType,
				Status:            sysTx.Status,
				Step:              sysTx.Step,
			})
		}

		errChan := make(chan error, len(batchParams))
		result := s.storage.TransferTransactions(repos.WithTx(tx)).BatchCreate(ctx, batchParams)
		defer func() {
			if clsErr := result.Close(); clsErr != nil {
				s.log.Errorw("close transfer transactions batch", "error", clsErr)
			}
		}()

		wg := sync.WaitGroup{}
		wg.Add(len(batchParams))

		result.Exec(func(_ int, err error) {
			defer wg.Done()
			errChan <- err
		})

		go func() {
			wg.Wait()
			close(errChan)
		}()

		for err := range errChan {
			if err != nil {
				return fmt.Errorf("batch create transfer system transactions: %w", err)
			}
		}

		return nil
	})
}

func (s *Service) createUnconfirmedTransaction(
	ctx context.Context,
	dto DepositWebhookDto,
	store *models.Store,
	wallet *models.Wallet,
	amount decimal.Decimal,
	usdAmount decimal.Decimal,
	opts repos.Option,
) (*models.UnconfirmedTransaction, error) {
	blockchain := string(*dto.Currency.Blockchain)
	uTransaction, err := s.unconfirmedTransactionsService.GetUnconfirmedByHash(ctx, dto.Hash, blockchain, opts)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("unconfirmed transaction found error: %w", err)
	}

	if err == nil {
		s.log.Info("unconfirmed transaction already exists")
		return uTransaction, nil
	}

	params := repo_unconfirmed_transactions.CreateParams{
		UserID:      store.UserID,
		StoreID:     uuid.NullUUID{UUID: store.ID, Valid: true},
		WalletID:    uuid.NullUUID{UUID: wallet.ID, Valid: true},
		CurrencyID:  dto.Currency.ID,
		Blockchain:  blockchain,
		TxHash:      dto.Hash,
		BcUniqKey:   &dto.TxUniqKey,
		Type:        models.TransactionsTypeDeposit,
		FromAddress: dto.FromAddress,
		ToAddress:   dto.ToAddress,
		Amount:      amount,
		AmountUsd:   decimal.NullDecimal{Decimal: usdAmount, Valid: true},
	}

	uTransaction, err = s.unconfirmedTransactionsService.CreateUnconfirmedTransaction(ctx, params, opts)
	if err != nil {
		return nil, fmt.Errorf("unconfirmed transaction creation: %w", err)
	}

	return uTransaction, nil
}

func (s *Service) checkUnconfirmed(dto WebhookDtoInterface) bool {
	return dto.GetStatus() == models.TransactionStatusWaitingConfirmations ||
		dto.GetStatus() == models.TransactionStatusInMempool
}

func (s *Service) enqueueAddressBalanceUpdate(ctx context.Context, dto WebhookDtoInterface, nativeTokenUpdateRequired bool, opts repos.Option) error {
	_, err := s.storage.UpdateBalanceQueue(opts).Create(ctx, repo_update_balance_queue.CreateParams{
		CurrencyID:               dto.GetCurrency().ID,
		Address:                  dto.GetAddressForUpdateBalance(),
		NativeTokenBalanceUpdate: nativeTokenUpdateRequired,
	})
	if err != nil {
		return fmt.Errorf("enqueue address balance: %w", err)
	}

	return nil
}
