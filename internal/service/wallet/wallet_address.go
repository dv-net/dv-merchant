package wallet

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/dv-net/dv-merchant/pkg/dbutils/pgerror"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"
	"github.com/gocarina/gocsv"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Service) RefreshWalletAddress(ctx context.Context, walletID uuid.UUID, address string) error {
	wallet, err := s.storage.Wallets().GetById(ctx, walletID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	str, err := s.storage.Stores().GetByID(ctx, wallet.StoreID)
	if err != nil {
		return fmt.Errorf("store not found: %w", err)
	}

	usr, err := s.storage.Users().GetByID(ctx, str.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	return s.MarkAddressDirty(ctx, usr, address)
}

// MarkAddressDirty TODO refactoring add marking address on processing and change logic get new wallets
func (s *Service) MarkAddressDirty(ctx context.Context, usr *models.User, address string) error {
	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		hasTransactions, err := s.storage.Transactions().HasTransactionsByAddress(ctx, address)
		if err != nil {
			return fmt.Errorf("failed to check transactions for address: %w", err)
		}
		if !hasTransactions {
			return ErrAddressHasNoTransactions
		}

		wa, err := s.storage.WalletAddresses(repos.WithTx(tx)).MarkAddressDirty(ctx, address, usr.ID)
		if err != nil {
			parsedErr := pgerror.ParseError(err)
			s.logger.Debug("error mark address is dirty", parsedErr)
			return parsedErr
		}
		for _, walletAddress := range wa {
			if err := s.processingService.MarkDirtyHotWallet(ctx, usr.ProcessingOwnerID.UUID, walletAddress.Blockchain, address); err != nil {
				return fmt.Errorf("failed to mark dirty in processing: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) LoadPrivateAddresses(ctx context.Context, dto LoadPrivateKeyDTO) (*bytes.Buffer, error) {
	data, err := s.processingService.GetOwnerHotWalletKeys(ctx, dto.User, dto.Otp, processing.GetOwnerHotWalletKeysParams{
		WalletAddressIDs:           dto.WalletAddressIDs,
		ExcludedWalletAddressesIDs: dto.ExcludedWalletAddressesIDs,
	})
	if err != nil {
		return nil, err
	}

	bb := new(bytes.Buffer)
	switch dto.FileType {
	case "json":
		j, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		if _, err := bb.Write(j); err != nil {
			return nil, err
		}
	case "csv":
		keys := make([]*HotKeyCsv, 0, len(data.Entries))
		for _, entry := range data.Entries {
			for _, item := range entry.Items {
				keys = append(keys, &HotKeyCsv{
					Blockchain: entry.Name,
					PublicKey:  item.PublicKey,
					PrivateKey: item.PrivateKey,
					Address:    item.Address,
				})
			}
		}
		if err := gocsv.Marshal(keys, bb); err != nil {
			return nil, err
		}
	case "txt":
		for _, entry := range data.Entries {
			for _, item := range entry.Items {
				if _, err := bb.WriteString(item.PrivateKey + "\n"); err != nil {
					return nil, err
				}
			}
		}
	}

	for _, item := range data.AllSelectedWallets {
		if err := s.logProcessingLoadAddressPrivateKey(ctx, item.Address, item.WalletAddressesID, dto.IP); err != nil {
			s.logger.Errorw("error store DB log for load private key", "error", err)
		}
	}

	return bb, nil
}

// getOrCreateWalletAddress returns existing or creates a new wallet address
func (s *Service) getOrCreateWalletAddress(
	ctx context.Context,
	dbTx pgx.Tx,
	storeOwner *models.User,
	wallet *models.Wallet,
	c *models.Currency,
) (*models.WalletAddress, error) {
	if c.IsFiat {
		return nil, fmt.Errorf("failed to create address for fiat currency")
	}

	if c.Blockchain == nil || *c.Blockchain == "" {
		return nil, fmt.Errorf("blockchain is not set for currency %s", c.ID)
	}

	key := wallet.ID.String() + ":" + c.ID
	muIface, _ := s.muMap.LoadOrStore(key, &sync.Mutex{})
	mu, ok := muIface.(*sync.Mutex)
	if !ok {
		return nil, fmt.Errorf("failed to get mutex for wallet address creation")
	}

	mu.Lock()
	defer func() {
		mu.Unlock()
		time.AfterFunc(100*time.Millisecond, func() {
			s.muMap.Delete(key)
		})
	}()

	const maxRetries = 5
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		walletAddress, err := s.storage.WalletAddresses(repos.WithTx(dbTx)).GetByWalletIDAndCurrencyID(ctx, wallet.ID, c.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("failed to get wallet address: %w", err)
		}

		if err == nil {
			if walletAddress.Dirty {
				return s.createNewWalletAddress(ctx, dbTx, storeOwner, wallet, c, walletAddress)
			}

			if logErr := s.logProcessingAddressReceived(ctx, walletAddress, pgtypeutils.DecodeText(wallet.IpAddress)); logErr != nil {
				s.logger.Errorw("failed create log to process processing addresses", "error", logErr)
			}

			return walletAddress, nil
		}

		addr, err := s.createNewWalletAddress(ctx, dbTx, storeOwner, wallet, c, nil)
		if err != nil {
			if isDuplicateErr(err) {
				s.logger.Debugw("duplicate detected, retrying", "attempt", attempt, "wallet_id", wallet.ID, "currency_id", c.ID)
				time.Sleep(time.Duration(attempt) * 50 * time.Millisecond)
				continue
			}

			lastErr = err
			s.logger.Warnw("error creating wallet address", "attempt", attempt, "error", err, "wallet_id", wallet.ID, "currency_id", c.ID)
			time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
			continue
		}

		return addr, nil
	}

	walletAddress, err := s.storage.WalletAddresses(repos.WithTx(dbTx)).GetByWalletIDAndCurrencyID(ctx, wallet.ID, c.ID)
	if err == nil {
		s.logger.Warnw("address found after retries", "wallet_id", wallet.ID, "currency_id", c.ID)
		return walletAddress, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to get or create wallet address after %d retries: %w", maxRetries, lastErr)
	}

	return nil, fmt.Errorf("failed to get or create wallet address after %d retries: address not found and no error recorded", maxRetries)
}

func (s *Service) createNewWalletAddress(
	ctx context.Context,
	dbTx pgx.Tx,
	storeOwner *models.User,
	wallet *models.Wallet,
	c *models.Currency,
	oldWalletAddress *models.WalletAddress,
) (*models.WalletAddress, error) {
	params := processing.CreateOwnerHotWalletParams{
		OwnerID:    storeOwner.ProcessingOwnerID.UUID,
		CustomerID: wallet.ID.String(),
		Blockchain: *c.Blockchain,
	}

	switch *c.Blockchain {
	case models.BlockchainBitcoin:
		params.BitcoinAddressType = util.Pointer(processing.ConvertToBitcoinAddressType(s.cfg.Blockchain.Bitcoin.AddressType))
	case models.BlockchainLitecoin:
		params.LitecoinAddressType = util.Pointer(processing.ConvertToLitecoinAddressType(s.cfg.Blockchain.Litecoin.AddressType))
	}

	newWallet, err := s.processingService.CreateOwnerHotWallet(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create new hot wallet: %w", err)
	}

	if oldWalletAddress != nil && newWallet.Address == oldWalletAddress.Address {
		return nil, fmt.Errorf("failed to create new wallet address: new address is the same as the old one")
	}

	walletAddress, err := s.storage.WalletAddresses(repos.WithTx(dbTx)).Create(ctx, repo_wallet_addresses.CreateParams{
		WalletID:   wallet.ID,
		UserID:     storeOwner.ID,
		CurrencyID: c.ID,
		Blockchain: *c.Blockchain,
		Address:    newWallet.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create new wallet address: %w", err)
	}

	return walletAddress, nil
}

func isDuplicateErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "unique constraint") ||
		strings.Contains(msg, "already exists")
}
