package address_book

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_address_book"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_withdrawal_wallet_addresses"

	"github.com/jackc/pgx/v5"
)

func (s *Service) deleteUniversalAddress(ctx context.Context, usr *models.User, address string, blockchain models.Blockchain, deleteWithdrawalRule bool) error {
	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		entries, err := s.storage.UserAddressBook().GetByUserAddressAndBlockchain(ctx, repo_user_address_book.GetByUserAddressAndBlockchainParams{
			UserID:     usr.ID,
			Address:    address,
			Blockchain: &blockchain,
			Type:       models.AddressBookTypeUniversal,
		})
		if err != nil {
			return fmt.Errorf("failed to get address entries: %w", err)
		}

		// If no entries found, return early
		if len(entries) == 0 {
			return nil
		}

		for _, entry := range entries {
			if err := s.storage.UserAddressBook(repos.WithTx(tx)).SoftDelete(ctx, entry.ID); err != nil {
				return fmt.Errorf("failed to delete address entry %s: %w", entry.ID, err)
			}

			if deleteWithdrawalRule {
				if err := s.cleanupWithdrawalRule(ctx, usr, entry, tx); err != nil {
					return fmt.Errorf("failed to cleanup withdrawal rule for entry %s: %w", entry.ID, err)
				}
			}

			s.logger.Info("Deleted universal address entry",
				"user_id", usr.ID,
				"address", address,
				"currency", entry.CurrencyID,
				"blockchain", blockchain.String())
		}

		return nil
	})
}

func (s *Service) deleteEVMAddress(ctx context.Context, usr *models.User, address string, deleteWithdrawalRule bool) error {
	var deletedCount int

	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		entries, err := s.storage.UserAddressBook().GetByUserAndAddressAllCurrencies(ctx, repo_user_address_book.GetByUserAndAddressAllCurrenciesParams{
			UserID:  usr.ID,
			Address: address,
			Type:    models.AddressBookTypeEVM,
		})
		if err != nil {
			return fmt.Errorf("failed to get address entries: %w", err)
		}

		var evmEntries []*models.UserAddressBook
		for _, entry := range entries {
			if entry.Blockchain != nil && entry.Blockchain.IsEVMLike() {
				evmEntries = append(evmEntries, entry)
			}
		}

		if len(evmEntries) == 0 {
			s.logger.Warn("No EVM address entries found for deletion",
				"user_id", usr.ID,
				"address", address)
			return fmt.Errorf("no EVM address entries found for user %s, address %s", usr.ID, address)
		}

		for _, entry := range evmEntries {
			if err := s.storage.UserAddressBook(repos.WithTx(tx)).SoftDelete(ctx, entry.ID); err != nil {
				return fmt.Errorf("failed to delete EVM address entry %s: %w", entry.ID, err)
			}

			if deleteWithdrawalRule {
				if err := s.cleanupWithdrawalRule(ctx, usr, entry, tx); err != nil {
					return fmt.Errorf("failed to cleanup withdrawal rule for entry %s: %w", entry.ID, err)
				}
			}

			s.logger.Infow("Deleted EVM address entry",
				"user_id", usr.ID,
				"address", address,
				"currency", entry.CurrencyID,
				"blockchain", entry.Blockchain.String())
		}

		deletedCount = len(evmEntries)
		return nil
	})

	if err != nil {
		return err
	}

	s.logger.Infow("Deleted EVM address and all associated entries",
		"user_id", usr.ID,
		"address", address,
		"entries_count", deletedCount)

	return nil
}

func (s *Service) cleanupWithdrawalRule(ctx context.Context, usr *models.User, addressEntry *models.UserAddressBook, tx pgx.Tx) error {
	withdrawalWallet, err := s.withdrawalWalletService.GetWithdrawalWalletsByCurrencyID(ctx, usr, addressEntry.CurrencyID, repos.WithTx(tx))
	if err != nil {
		s.logger.Warn("Withdrawal wallet not found for cleanup",
			"currency", addressEntry.CurrencyID,
			"user_id", addressEntry.UserID)
		return nil //nolint:nilerr
	}

	withdrawalAddress, err := s.storage.WithdrawalWalletAddresses().GetByAddress(ctx, repo_withdrawal_wallet_addresses.GetByAddressParams{
		WithdrawalWalletID: withdrawalWallet.ID,
		Address:            addressEntry.Address,
	})
	if err != nil {
		s.logger.Warn("Withdrawal wallet address not found for cleanup",
			"address", addressEntry.Address,
			"currency", addressEntry.CurrencyID)
		return nil //nolint:nilerr
	}

	err = s.storage.WithdrawalWalletAddresses(repos.WithTx(tx)).SoftDelete(ctx, withdrawalAddress.ID)
	if err != nil {
		return fmt.Errorf("failed to delete withdrawal wallet address: %w", err)
	}

	s.logger.Info("Cleaned up withdrawal rule for address book entry",
		"address", addressEntry.Address,
		"currency", addressEntry.CurrencyID,
		"user_id", addressEntry.UserID)

	return nil
}
