package address_book

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_address_book"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_withdrawal_wallet_addresses"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) createSingleAddress(ctx context.Context, params CreateAddressDTO) (*models.UserAddressBook, error) {
	var address *models.UserAddressBook

	user, err := s.storage.Users().GetByID(ctx, params.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	err = repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		blockchain := params.Blockchain
		if blockchain == nil {
			currency, err := s.currencyService.GetCurrencyByID(ctx, params.CurrencyID)
			if err != nil {
				return fmt.Errorf("failed to get currency: %w", err)
			}
			if currency.Blockchain == nil {
				return fmt.Errorf("currency %s does not have a blockchain assigned", params.CurrencyID)
			}
			blockchain = currency.Blockchain
		}

		exists, err := s.storage.UserAddressBook(repos.WithTx(tx)).CheckExists(ctx, repo_user_address_book.CheckExistsParams{
			UserID:     params.UserID,
			Address:    params.Address,
			CurrencyID: params.CurrencyID,
			Type:       models.AddressBookTypeSimple,
		})
		if err != nil {
			return fmt.Errorf("failed to check if address exists: %w", err)
		}

		if exists { //nolint:nestif
			var err error
			address, err = s.storage.UserAddressBook(repos.WithTx(tx)).GetByUserAndAddress(ctx, repo_user_address_book.GetByUserAndAddressParams{
				UserID:     params.UserID,
				Address:    params.Address,
				CurrencyID: params.CurrencyID,
				Type:       models.AddressBookTypeSimple,
			})
			if err != nil {
				return fmt.Errorf("failed to get existing address entry: %w", err)
			}

			// If address already exists and is active return error
			if !address.DeletedAt.Valid {
				return fmt.Errorf("address book entry already exists and is active")
			}

			// If address exists but is soft-deleted, restore it
			address, err = s.restoreAddressEntry(ctx, params, tx)
			if err != nil {
				return err
			}

			// Check if we need to create/restore a withdrawal rule
			if params.CreateWithdrawalRule {
				if err := s.restoreWithdrawalRule(ctx, address, user, params.TOTP, tx); err != nil {
					return fmt.Errorf("failed to create withdrawal rule: %w", err)
				}
			}
		} else {
			address, err = s.createNewAddressEntry(ctx, params, blockchain, tx)
			if err != nil {
				return err
			}
			// Check if we need to create/restore a withdrawal rule
			if params.CreateWithdrawalRule {
				if err := s.restoreWithdrawalRule(ctx, address, user, params.TOTP, tx); err != nil {
					return fmt.Errorf("failed to create withdrawal rule: %w", err)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return address, nil
}

func (s *Service) createUniversalAddress(ctx context.Context, params CreateAddressDTO) (*models.UserAddressBook, error) {
	if params.Blockchain == nil {
		return nil, fmt.Errorf("blockchain is required for universal addresses")
	}

	var user *models.User
	// Fetch user if TOTP is provided (needed for withdrawal rule creation)
	if params.TOTP != "" {
		var err error
		user, err = s.storage.Users().GetByID(ctx, params.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
	}

	currencies, err := s.currencyService.GetCurrenciesByBlockchain(ctx, *params.Blockchain)
	if err != nil {
		return nil, fmt.Errorf("failed to get currencies for blockchain %s: %w", params.Blockchain.String(), err)
	}

	if len(currencies) == 0 {
		return nil, fmt.Errorf("no currencies found for blockchain %s", params.Blockchain.String())
	}

	var firstAddress *models.UserAddressBook

	err = repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		for _, currency := range currencies {
			currencyParams := params
			currencyParams.CurrencyID = currency.ID

			address, err := s.handleAddressCreation(
				ctx,
				currencyParams,
				models.AddressBookTypeUniversal,
				currency,
				params.Blockchain,
				user,
				params.TOTP,
				tx,
				params.CreateWithdrawalRule,
			)
			if err != nil {
				return err
			}

			if firstAddress == nil {
				firstAddress = address
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if firstAddress == nil {
		return nil, fmt.Errorf("no address entries were created (all already existed)")
	}

	s.logger.Info("Created universal address for blockchain",
		"user_id", params.UserID,
		"address", params.Address,
		"blockchain", params.Blockchain.String(),
		"currencies_count", len(currencies))

	return firstAddress, nil
}

func (s *Service) createEVMAddress(ctx context.Context, params CreateAddressDTO) (*models.UserAddressBook, error) {
	var user *models.User
	// Fetch user if TOTP is provided (needed for withdrawal rule creation)
	if params.TOTP != "" {
		var err error
		user, err = s.storage.Users().GetByID(ctx, params.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
	}

	evmBlockchains := []models.Blockchain{
		models.BlockchainEthereum,
		models.BlockchainBinanceSmartChain,
		models.BlockchainPolygon,
		models.BlockchainArbitrum,
	}

	var firstAddress *models.UserAddressBook
	var totalCurrencyCount int

	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		for _, blockchain := range evmBlockchains {
			currencies, err := s.currencyService.GetCurrenciesByBlockchain(ctx, blockchain)
			if err != nil {
				return fmt.Errorf("failed to get currencies for EVM blockchain %s: %w", blockchain.String(), err)
			}

			if len(currencies) == 0 {
				s.logger.Warn("No currencies found for EVM blockchain, skipping",
					"blockchain", blockchain.String())
				continue
			}

			for _, currency := range currencies {
				evmParams := params
				evmParams.Blockchain = &blockchain
				evmParams.CurrencyID = currency.ID

				address, err := s.handleAddressCreation(ctx, evmParams, models.AddressBookTypeEVM,
					currency,
					&blockchain,
					user,
					params.TOTP,
					tx,
					params.CreateWithdrawalRule,
				)
				if err != nil {
					return err
				}

				if firstAddress == nil {
					firstAddress = address
				}

				totalCurrencyCount++
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if firstAddress == nil {
		return nil, fmt.Errorf("no EVM address entries were created (all already existed)")
	}

	s.logger.Info("Created EVM address across all EVM blockchains",
		"user_id", params.UserID,
		"address", params.Address,
		"total_currencies", totalCurrencyCount,
		"evm_blockchains", len(evmBlockchains))

	return firstAddress, nil
}

func (s *Service) restoreAddressEntry(ctx context.Context, params CreateAddressDTO, tx pgx.Tx) (*models.UserAddressBook, error) {
	restoreParams := repo_user_address_book.RestoreFromTrashParams{
		UserID:     params.UserID,
		Address:    params.Address,
		CurrencyID: params.CurrencyID,
	}

	restoreParams.Type = models.AddressBookTypeSimple
	if params.EVM {
		restoreParams.Type = models.AddressBookTypeEVM
	}

	if params.Universal {
		restoreParams.Type = models.AddressBookTypeUniversal
	}

	if params.Name != nil {
		restoreParams.Name = pgtype.Text{String: *params.Name, Valid: true}
	}

	if params.Tag != nil {
		restoreParams.Tag = pgtype.Text{String: *params.Tag, Valid: true}
	}

	address, err := s.storage.UserAddressBook(repos.WithTx(tx)).RestoreFromTrash(ctx, restoreParams)
	if err != nil {
		return nil, fmt.Errorf("failed to restore address from trash: %w", err)
	}

	s.logger.Info("Restored soft-deleted address book entry",
		"user_id", params.UserID,
		"address", params.Address,
		"currency", params.CurrencyID)

	return address, nil
}

func (s *Service) createNewAddressEntry(ctx context.Context, params CreateAddressDTO, blockchain *models.Blockchain, tx pgx.Tx) (*models.UserAddressBook, error) {
	createParams := repo_user_address_book.CreateParams{
		UserID:     params.UserID,
		Address:    params.Address,
		CurrencyID: params.CurrencyID,
		Blockchain: blockchain,
	}

	createParams.Type = models.AddressBookTypeSimple

	if params.EVM {
		createParams.Type = models.AddressBookTypeEVM
	}

	if params.Universal {
		createParams.Type = models.AddressBookTypeUniversal
	}

	if params.Name != nil {
		createParams.Name = pgtype.Text{String: *params.Name, Valid: true}
	}

	if params.Tag != nil {
		createParams.Tag = pgtype.Text{String: *params.Tag, Valid: true}
	}

	address, err := s.storage.UserAddressBook(repos.WithTx(tx)).Create(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	return address, nil
}

func (s *Service) createWithdrawalRule(ctx context.Context, addressEntry *models.UserAddressBook, user *models.User, totp string, tx pgx.Tx) error {
	withdrawalWallet, err := s.withdrawalWalletService.GetWithdrawalWalletsByCurrencyID(ctx, addressEntry.UserID, addressEntry.CurrencyID)
	if err != nil {
		return fmt.Errorf("failed to get withdrawal wallet: %w", err)
	}

	createParams := repo_withdrawal_wallet_addresses.CreateParams{
		WithdrawalWalletID: withdrawalWallet.ID,
		Address:            addressEntry.Address,
	}

	if addressEntry.Name.Valid {
		createParams.Name = &addressEntry.Name.String
	}

	// Create the withdrawal wallet address within our transaction
	_, err = s.storage.WithdrawalWalletAddresses(repos.WithTx(tx)).Create(ctx, createParams)
	if err != nil {
		return fmt.Errorf("failed to create withdrawal wallet address: %w", err)
	}

	// Update processing whitelist with the new address
	if err := s.updateProcessingWhitelist(ctx, withdrawalWallet.ID, user, totp, tx); err != nil {
		return fmt.Errorf("failed to update processing whitelist: %w", err)
	}

	s.logger.Info("Created withdrawal rule for address book entry",
		"address", addressEntry.Address,
		"currency", addressEntry.CurrencyID,
		"user_id", addressEntry.UserID)

	return nil
}

func (s *Service) restoreWithdrawalRule(ctx context.Context, addressEntry *models.UserAddressBook, user *models.User, totp string, tx pgx.Tx) error {
	withdrawalWallet, err := s.withdrawalWalletService.GetWithdrawalWalletsByCurrencyID(ctx, addressEntry.UserID, addressEntry.CurrencyID)
	if err != nil {
		return fmt.Errorf("failed to get withdrawal wallet: %w", err)
	}

	withdrawalAddress, err := s.storage.WithdrawalWalletAddresses().GetByAddressWithTrashed(ctx, repo_withdrawal_wallet_addresses.GetByAddressWithTrashedParams{
		WithdrawalWalletID: withdrawalWallet.ID,
		Address:            addressEntry.Address,
	})
	if err != nil {
		s.logger.Info("No existing withdrawal rule found, creating new one",
			"address", addressEntry.Address,
			"currency", addressEntry.CurrencyID,
			"user_id", addressEntry.UserID)
		return s.createWithdrawalRule(ctx, addressEntry, user, totp, tx)
	}

	if withdrawalAddress.DeletedAt.Valid {
		var name *string
		if addressEntry.Name.Valid {
			name = &addressEntry.Name.String
		}

		_, err = s.storage.WithdrawalWalletAddresses(repos.WithTx(tx)).UpdateDeletedAddress(ctx, repo_withdrawal_wallet_addresses.UpdateDeletedAddressParams{
			ID:   withdrawalAddress.ID,
			Name: name,
		})
		if err != nil {
			return fmt.Errorf("failed to restore withdrawal wallet address: %w", err)
		}

		// Update processing whitelist after restoration
		if err := s.updateProcessingWhitelist(ctx, withdrawalWallet.ID, user, totp, tx); err != nil {
			return fmt.Errorf("failed to update processing whitelist: %w", err)
		}

		s.logger.Info("Restored soft-deleted withdrawal rule for address book entry",
			"address", addressEntry.Address,
			"currency", addressEntry.CurrencyID,
			"user_id", addressEntry.UserID)
	} else {
		s.logger.Info("Withdrawal rule already exists and is active",
			"address", addressEntry.Address,
			"currency", addressEntry.CurrencyID,
			"user_id", addressEntry.UserID)
	}

	return nil
}

func (s *Service) addWithdrawalRuleForSimpleAddress(ctx context.Context, userID uuid.UUID, addressID uuid.UUID, user *models.User, totp string) error {
	addressEntry, err := s.storage.UserAddressBook().GetByID(ctx, addressID)
	if err != nil {
		return fmt.Errorf("failed to get address entry: %w", err)
	}

	if addressEntry.UserID != userID {
		return fmt.Errorf("address entry does not belong to user")
	}

	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		return s.restoreWithdrawalRule(ctx, addressEntry, user, totp, tx)
	})
}

func (s *Service) addWithdrawalRulesForUniversalAddress(ctx context.Context, userID uuid.UUID, address string, blockchain models.Blockchain, user *models.User, totp string) error {
	entries, err := s.storage.UserAddressBook().GetByUserAddressAndBlockchain(ctx, repo_user_address_book.GetByUserAddressAndBlockchainParams{
		UserID:     userID,
		Address:    address,
		Blockchain: &blockchain,
	})
	if err != nil {
		return fmt.Errorf("failed to get universal address entries: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("no universal address entries found for user %s, address %s, blockchain %s", userID, address, blockchain.String())
	}

	var rulesCreated int

	err = repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		for _, entry := range entries {
			if err := s.restoreWithdrawalRule(ctx, entry, user, totp, tx); err != nil {
				return fmt.Errorf("failed to restore/create withdrawal rule for currency %s: %w", entry.CurrencyID, err)
			}

			rulesCreated++

			s.logger.Info("Added/restored withdrawal rule for universal address entry",
				"address", entry.Address,
				"currency", entry.CurrencyID,
				"blockchain", blockchain.String(),
				"user_id", userID)
		}

		return nil
	})

	if err != nil {
		return err
	}

	s.logger.Info("Added withdrawal rules for universal address",
		"user_id", userID,
		"address", address,
		"blockchain", blockchain.String(),
		"rules_created", rulesCreated,
		"total_entries", len(entries))

	return nil
}

func (s *Service) addWithdrawalRulesForEVMAddress(ctx context.Context, userID uuid.UUID, address string, user *models.User, totp string) error {
	entries, err := s.storage.UserAddressBook().GetByUserAndAddressAllCurrencies(ctx, repo_user_address_book.GetByUserAndAddressAllCurrenciesParams{
		UserID:  userID,
		Address: address,
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
		return fmt.Errorf("no EVM address entries found for user %s, address %s", userID, address)
	}

	var rulesCreated int

	err = repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		for _, entry := range evmEntries {
			if err := s.restoreWithdrawalRule(ctx, entry, user, totp, tx); err != nil {
				return fmt.Errorf("failed to restore/create withdrawal rule for currency %s: %w", entry.CurrencyID, err)
			}

			rulesCreated++

			s.logger.Info("Added/restored withdrawal rule for EVM address entry",
				"address", entry.Address,
				"currency", entry.CurrencyID,
				"blockchain", entry.Blockchain.String(),
				"user_id", userID)
		}

		return nil
	})

	if err != nil {
		return err
	}

	s.logger.Info("Added withdrawal rules for EVM address",
		"user_id", userID,
		"address", address,
		"rules_created", rulesCreated,
		"total_evm_entries", len(evmEntries))

	return nil
}

func (s *Service) handleAddressCreation(
	ctx context.Context,
	params CreateAddressDTO,
	addressBookType models.AddressBookType,
	currency *models.Currency,
	blockchain *models.Blockchain,
	user *models.User,
	totp string,
	tx pgx.Tx,
	createWithdrawalRule bool,
) (*models.UserAddressBook, error) {
	exists, err := s.storage.UserAddressBook(repos.WithTx(tx)).CheckExists(ctx, repo_user_address_book.CheckExistsParams{
		UserID:     params.UserID,
		Address:    params.Address,
		CurrencyID: params.CurrencyID,
		Type:       addressBookType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check if address exists for currency %s: %w", currency.ID, err)
	}

	var address *models.UserAddressBook

	if exists {
		address, err = s.handleExistingAddress(ctx, params, addressBookType, user, totp, tx, createWithdrawalRule)
		if err != nil {
			return nil, fmt.Errorf("failed to handle existing address: %w", err)
		}
	} else {
		address, err = s.handleNewAddress(ctx, params, blockchain, user, totp, tx, createWithdrawalRule)
		if err != nil {
			return nil, fmt.Errorf("failed to handle new address: %w", err)
		}
	}

	return address, nil
}

func (s *Service) handleExistingAddress(ctx context.Context, params CreateAddressDTO, addressBookType models.AddressBookType, user *models.User, totp string, tx pgx.Tx, createWithdrawalRule bool) (*models.UserAddressBook, error) {
	address, err := s.storage.UserAddressBook(repos.WithTx(tx)).GetByUserAndAddress(ctx, repo_user_address_book.GetByUserAndAddressParams{
		UserID:     params.UserID,
		Address:    params.Address,
		CurrencyID: params.CurrencyID,
		Type:       addressBookType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get existing address entry: %w", err)
	}

	if !address.DeletedAt.Valid {
		return nil, fmt.Errorf("address book entry already exists and is active")
	}

	address, err = s.restoreAddressEntry(ctx, params, tx)
	if err != nil {
		return nil, err
	}

	if createWithdrawalRule {
		if err := s.restoreWithdrawalRule(ctx, address, user, totp, tx); err != nil {
			return nil, fmt.Errorf("failed to create withdrawal rule for old entry: %w", err)
		}
	}
	return address, nil
}

func (s *Service) handleNewAddress(ctx context.Context, params CreateAddressDTO, blockchain *models.Blockchain, user *models.User, totp string, tx pgx.Tx, createWithdrawalRule bool) (*models.UserAddressBook, error) {
	address, err := s.createNewAddressEntry(ctx, params, blockchain, tx)
	if err != nil {
		return nil, err
	}
	if createWithdrawalRule {
		if err := s.restoreWithdrawalRule(ctx, address, user, totp, tx); err != nil {
			return nil, fmt.Errorf("failed to create withdrawal rule for new entry: %w", err)
		}
	}
	return address, nil
}

func (s *Service) updateProcessingWhitelist(ctx context.Context, withdrawalWalletID uuid.UUID, user *models.User, totp string, tx pgx.Tx) error {
	wallet, err := s.withdrawalWalletService.GetWalletByID(ctx, withdrawalWalletID)
	if err != nil {
		return err
	}

	addresses, err := s.storage.WithdrawalWalletAddresses(repos.WithTx(tx)).GetWithdrawalWalletsByBlockchain(ctx, repo_withdrawal_wallet_addresses.GetWithdrawalWalletsByBlockchainParams{
		Blockchain: wallet.Blockchain,
		UserID:     wallet.UserID,
	})
	if err != nil {
		return fmt.Errorf("failed to get withdrawal wallet addresses: %w", err)
	}

	params := processing.AttachOwnerColdWalletsParams{
		OwnerID:    user.ProcessingOwnerID.UUID,
		Blockchain: wallet.Blockchain,
		Addresses:  addresses,
		TOTP:       totp,
	}

	err = s.processingWalletService.AttachOwnerColdWallets(ctx, params)
	if err != nil {
		return fmt.Errorf("update withdrawal wallet processing whitelist: %w", err)
	}
	return nil
}
