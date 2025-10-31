package wallet

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/constant"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses"
	"github.com/google/uuid"
)

type IWalletAddressLifecycle interface {
	// Reserve mark wallet at "reserved"
	Reserve(ctx context.Context, walletAddress *models.WalletAddress, opts ...repos.Option) error
	// Release releases the wallet and returns it to the pool (status = released → available).
	Release(ctx context.Context, walletAddress *models.WalletAddress, opts ...repos.Option) error
	// ReleaseByAccountID releases all wallets by account ID
	ReleaseByAccountID(ctx context.Context, accountID uuid.UUID, opts ...repos.Option) error
}

func (s *Service) Reserve(ctx context.Context, walletAddress *models.WalletAddress, opts ...repos.Option) error {
	err := s.storage.WalletAddresses(opts...).UpdateStatus(ctx,
		constant.WalletStatusReserved, walletAddress.ID)
	if err != nil {
		return err
	}

	s.logger.Debugw("wallet address reserved",
		"wallet_address_id", walletAddress.ID,
		"wallet_address_status", walletAddress.Status)

	err = s.logWalletStatusChanged(ctx, walletAddress, walletAddress.Status.String(), constant.WalletStatusReserved.String())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Release(ctx context.Context, walletAddress *models.WalletAddress, opts ...repos.Option) error {
	err := s.storage.WalletAddresses(opts...).UpdateStatus(ctx,
		constant.WalletStatusAvailable,
		walletAddress.ID)
	if err != nil {
		return err
	}
	s.logger.Debugw("wallet address reserved",
		"wallet_address_id", walletAddress.ID,
		"wallet_address_status", walletAddress.Status)

	err = s.logWalletStatusChanged(ctx, walletAddress, walletAddress.Status.String(), constant.WalletStatusAvailable.String())
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ReleaseByAccountID(ctx context.Context, accountID uuid.UUID, opts ...repos.Option) error {
	err := s.storage.WalletAddresses(opts...).UpdateStatusByAccountID(ctx, repo_wallet_addresses.UpdateStatusByAccountIDParams{
		AccountID:   uuid.NullUUID{UUID: accountID, Valid: true},
		AccountType: constant.RotateAddress.String(),
		Status:      constant.WalletStatusAvailable,
	})
	if err != nil {
		return err
	}
	s.logger.Debugw("wallet address released by account id",
		"account id", accountID.ID,
		"wallet_address_status", constant.WalletStatusAvailable)
	return nil
}
