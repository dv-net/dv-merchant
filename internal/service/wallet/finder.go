package wallet

import (
	"context"
	"errors"

	"github.com/dv-net/dv-merchant/internal/constant"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type IWalletAddressFinder interface {
	// GetAvailable get first available currency by id
	GetAvailable(ctx context.Context, usr *models.User, store *models.Store, cur *models.Currency) (*models.WalletAddress, error)
	// GetByAccountID return wallet by account id and type
	GetByAccountID(ctx context.Context, id uuid.UUID, wType constant.WalletAddressType) ([]*models.WalletAddress, error)
	// GetByAccountIDAndCurrency return wallet by account id, type and currency
	GetByAccountAndCurrency(ctx context.Context, id uuid.UUID, currencyID string) (*models.WalletAddress, error)
}

func (s *Service) GetAvailable(ctx context.Context, usr *models.User, store *models.Store, cur *models.Currency) (*models.WalletAddress, error) {
	wAddress, err := s.storage.WalletAddresses().GetAvailableAddress(ctx, cur.ID, usr.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		walletAddress, err := s.AddFromProcessing(ctx, usr, store, cur)
		if err != nil {
			return nil, err
		}
		return walletAddress, nil
	}

	return wAddress, nil
}

func (s *Service) GetByAccountID(ctx context.Context, id uuid.UUID, wType constant.WalletAddressType) ([]*models.WalletAddress, error) {
	wAddress, err := s.storage.WalletAddresses().GetWalletAddressByTypeAndID(ctx, uuid.NullUUID{UUID: id, Valid: true}, wType.String())
	if err != nil {
		return nil, err
	}
	return wAddress, nil
}

func (s *Service) GetByAccountAndCurrency(ctx context.Context, id uuid.UUID, currencyID string) (*models.WalletAddress, error) {
	wAddress, err := s.storage.WalletAddresses().GetByWalletIDAndCurrencyID(ctx, uuid.NullUUID{UUID: id, Valid: true}, currencyID)
	if err != nil {
		return nil, err
	}
	return wAddress, nil
}
