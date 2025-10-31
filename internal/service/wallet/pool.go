package wallet

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/constant"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/google/uuid"
)

type IWalletAddressPoolManager interface {
	// AddFromProcessing wallet in pool
	AddFromProcessing(ctx context.Context, usr *models.User, store *models.Store, cur *models.Currency) (*models.WalletAddress, error)
}

func (s *Service) AddFromProcessing(ctx context.Context, usr *models.User, store *models.Store, c *models.Currency) (*models.WalletAddress, error) {
	walletAddressID := uuid.New()
	params := processing.CreateOwnerHotWalletParams{
		OwnerID:    usr.ProcessingOwnerID.UUID,
		CustomerID: walletAddressID.String(),
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
	if !c.Blockchain.IsBitcoinLike() {
		currencies, err := s.currencyService.GetCurrenciesByBlockchain(ctx, *c.Blockchain)
		if err != nil {
			return nil, fmt.Errorf("failed to get currencies for blockchain %s: %w", *c.Blockchain, err)
		}

		var created *models.WalletAddress
		for _, currency := range currencies {
			wAddr, err := s.storage.WalletAddresses().Create(ctx, repo_wallet_addresses.CreateParams{
				AccountID:   uuid.NullUUID{UUID: walletAddressID, Valid: true},
				AccountType: constant.RotateAddress.String(),
				UserID:      usr.ID,
				CurrencyID:  currency.ID,
				Blockchain:  *c.Blockchain,
				Address:     newWallet.Address,
				Status:      constant.WalletStatusAvailable,
				StoreID:     uuid.NullUUID{UUID: store.ID, Valid: true},
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create wallet address for %s: %w", currency.ID, err)
			}

			if currency.ID == c.ID {
				created = wAddr
			}
		}

		if created == nil {
			return nil, fmt.Errorf("wallet for currency %s not found after creation", c.ID)
		}

		return created, nil
	}
	// no evm
	walletAddress, err := s.storage.WalletAddresses().Create(ctx, repo_wallet_addresses.CreateParams{
		AccountID:   uuid.NullUUID{UUID: walletAddressID, Valid: true},
		AccountType: constant.RotateAddress.String(),
		UserID:      usr.ID,
		CurrencyID:  c.ID,
		Blockchain:  *c.Blockchain,
		Address:     newWallet.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save wallet address: %w", err)
	}

	return walletAddress, nil
}
