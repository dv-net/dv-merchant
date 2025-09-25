package withdrawal_wallet

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/dv-net/dv-merchant/internal/tools/apierror"

	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_multi_withdrawal_rules"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_withdrawal_wallet_addresses"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/currency"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_withdrawal_wallets"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type IWithdrawalWalletService interface {
	GetWalletByID(ctx context.Context, id uuid.UUID) (*models.WithdrawalWallet, error)
	GetWithdrawalWallets(ctx context.Context, userID uuid.UUID) ([]*WithdrawalWithAddress, error)
	GetWithdrawalWalletsByCurrencyID(ctx context.Context, userID uuid.UUID, currencyID string, opts ...repos.Option) (*WithdrawalWithAddress, error)
	UpdateWithdrawalRules(ctx context.Context, dto UpdateRulesDTO) error

	GetAddressByID(ctx context.Context, id uuid.UUID) (*models.WithdrawalWalletAddress, error)
	CreateWithdrawalWalletAddress(ctx context.Context, params repo_withdrawal_wallet_addresses.CreateParams, user *models.User, totp string) (*models.WithdrawalWalletAddress, error)
	BatchCreateOrUpdateWallet(ctx context.Context, user *models.User, dto UpdateAddressesListDTO) error
	UpdateAddress(ctx context.Context, params repo_withdrawal_wallet_addresses.UpdateDeletedAddressParams, opts ...repos.Option) (*models.WithdrawalWalletAddress, error)
	DeleteAddress(ctx context.Context, walletAddressID uuid.UUID, user *models.User, totp string) error
	BatchDelete(ctx context.Context, withdrawalWalletID uuid.UUID, addressesIDs []uuid.UUID, user *models.User, totp string) error
}

type Service struct {
	storage                 storage.IStorage
	logger                  logger.Logger
	currencyService         currency.ICurrency
	currConvService         currconv.ICurrencyConvertor
	processingWalletService processing.IProcessingWallet
}

func New(
	storage storage.IStorage,
	logger logger.Logger,
	currencyService currency.ICurrency,
	currConvService currconv.ICurrencyConvertor,
	processingWalletService processing.IProcessingWallet,
) *Service {
	return &Service{
		storage:                 storage,
		logger:                  logger,
		currencyService:         currencyService,
		currConvService:         currConvService,
		processingWalletService: processingWalletService,
	}
}

func (s Service) GetWithdrawalWallets(ctx context.Context, userID uuid.UUID) ([]*WithdrawalWithAddress, error) {
	wallets, err := s.storage.WithdrawalWallets().GetWithdrawalWallets(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawal wallets: %w", err)
	}

	availableCurrencies, err := s.currencyService.GetCurrenciesEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get currencies: %w", err)
	}

	walletMap := make(map[string]*models.WithdrawalWallet, len(wallets))
	for _, wallet := range wallets {
		walletMap[wallet.CurrencyID] = wallet
	}

	data := make([]*WithdrawalWithAddress, 0, len(availableCurrencies))

	for _, c := range availableCurrencies {
		mCurrency := models.CurrencyShort{
			ID:            c.ID,
			Code:          c.Code,
			Precision:     c.Precision,
			Name:          c.Name,
			Blockchain:    c.Blockchain,
			IsBitcoinLike: c.Blockchain.IsBitcoinLike(),
			IsEVMLike:     c.Blockchain.IsEVMLike(),
		}

		wallet, exists := walletMap[c.ID]
		if !exists {
			wallet, err = s.createWithdrawalWallet(ctx, c, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to create withdrawal wallet for currency %s: %w", c.ID, err)
			}
		}

		withdrawalWithAddress, err := s.buildWithdrawalWithAddress(ctx, wallet, &mCurrency)
		if err != nil {
			return nil, err
		}

		data = append(data, withdrawalWithAddress)
	}

	return data, nil
}

func (s Service) GetWithdrawalWalletsByCurrencyID(ctx context.Context, userID uuid.UUID, currencyID string, opts ...repos.Option) (*WithdrawalWithAddress, error) {
	c, err := s.currencyService.GetCurrencyByID(ctx, currencyID)
	if err != nil {
		return nil, fmt.Errorf("currency not found: %s: %w", currencyID, err)
	}

	wallet, err := s.storage.WithdrawalWallets(opts...).GetWithdrawalWalletByCurrency(ctx, repo_withdrawal_wallets.GetWithdrawalWalletByCurrencyParams{
		UserID:     userID,
		CurrencyID: currencyID,
	})

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get withdrawal wallets by currency %s: %w", currencyID, err)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		wallet, err = s.createWithdrawalWallet(ctx, c, userID, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create withdrawal wallet for currency %s: %w", currencyID, err)
		}
	}

	mCurrency := models.CurrencyShort{
		ID:            c.ID,
		Code:          c.Code,
		Precision:     c.Precision,
		Name:          c.Name,
		Blockchain:    c.Blockchain,
		IsBitcoinLike: c.Blockchain.IsBitcoinLike(),
		IsEVMLike:     c.Blockchain.IsEVMLike(),
	}

	return s.buildWithdrawalWithAddress(ctx, wallet, &mCurrency, opts...)
}

func (s Service) UpdateWithdrawalRules(ctx context.Context, dto UpdateRulesDTO) error {
	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		withdrawalWallet, err := s.storage.WithdrawalWallets(repos.WithTx(tx)).Update(
			ctx,
			repo_withdrawal_wallets.UpdateParams{
				WithdrawalEnabled:       dto.WithdrawalEnabled,
				WithdrawalMinBalance:    dto.WithdrawalMinBalance,
				WithdrawalMinBalanceUsd: dto.WithdrawalMinBalanceUsd,
				WithdrawalInterval:      dto.WithdrawalInterval,
				CurrencyID:              dto.Currency.ID,
				UserID:                  dto.UserID,
			},
		)
		if err != nil {
			return err
		}

		if dto.MultiWithdrawal == nil {
			return s.storage.MultiWithdrawalRules().RemoveByWalletID(ctx, withdrawalWallet.ID)
		}

		if !dto.Currency.Blockchain.IsBitcoinLike() {
			return fmt.Errorf("multi withdrawal rules supported only for bitcoin-like chains")
		}

		multiWithdrawalsParam := repo_multi_withdrawal_rules.CreateOrUpdateParams{
			WithdrawalWalletID: withdrawalWallet.ID,
			Mode:               dto.MultiWithdrawal.Mode,
		}

		if dto.MultiWithdrawal.ManualAddress != nil && *dto.MultiWithdrawal.ManualAddress != "" {
			exists, err := s.storage.WithdrawalWalletAddresses(repos.WithTx(tx)).CheckAddressExists(
				ctx,
				repo_withdrawal_wallet_addresses.CheckAddressExistsParams{
					WithdrawalWalletID: withdrawalWallet.ID,
					Address:            *dto.MultiWithdrawal.ManualAddress,
				},
			)

			if !exists || err != nil {
				return fmt.Errorf("failed to find cold wallet in approved addresses")
			}

			multiWithdrawalsParam.ManualAddress = pgtype.Text{Valid: true, String: *dto.MultiWithdrawal.ManualAddress}
		}

		return s.storage.MultiWithdrawalRules(repos.WithTx(tx)).CreateOrUpdate(ctx, multiWithdrawalsParam)
	})
}

func (s Service) CreateWithdrawalWalletAddress(ctx context.Context, params repo_withdrawal_wallet_addresses.CreateParams, user *models.User, totp string) (*models.WithdrawalWalletAddress, error) {
	var updatedAddress *models.WithdrawalWalletAddress

	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		address, err := s.storage.WithdrawalWalletAddresses(repos.WithTx(tx)).GetByAddressWithTrashed(ctx, repo_withdrawal_wallet_addresses.GetByAddressWithTrashedParams{
			WithdrawalWalletID: params.WithdrawalWalletID,
			Address:            params.Address,
		})
		if err != nil { //nolint:nestif
			if errors.Is(err, pgx.ErrNoRows) {
				newAddress, err := s.storage.WithdrawalWalletAddresses(repos.WithTx(tx)).Create(ctx, params)
				if err != nil {
					return fmt.Errorf("failed to create withdrawal wallet address: %w", err)
				}
				updatedAddress = newAddress

				if err := s.updateProcessingWithdrawalWalletWhitelist(ctx, params.WithdrawalWalletID, user, totp, tx); err != nil {
					return apierror.New().AddError(fmt.Errorf("update withdrawal wallet address %s: %w", newAddress.Address, err)).SetHttpCode(fiber.StatusBadRequest)
				}
				return nil
			}
			return fmt.Errorf("failed to get withdrawal wallet address: %w", err)
		}
		// update address if him deleted
		updatedAddress, err = s.UpdateAddress(ctx, repo_withdrawal_wallet_addresses.UpdateDeletedAddressParams{
			Name: params.Name,
			ID:   address.ID,
		}, repos.WithTx(tx))
		if err != nil {
			return err
		}

		err = s.updateProcessingWithdrawalWalletWhitelist(ctx, params.WithdrawalWalletID, user, totp, tx)
		if err != nil {
			return apierror.New().AddError(err).SetHttpCode(fiber.StatusBadRequest)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return updatedAddress, nil
}

func (s Service) DeleteAddress(ctx context.Context, walletAddressID uuid.UUID, user *models.User, totp string) error {
	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		walletAddress, err := s.GetAddressByID(ctx, walletAddressID)
		if err != nil {
			return err
		}

		err = s.storage.WithdrawalWalletAddresses(repos.WithTx(tx)).SoftDelete(ctx, walletAddressID)
		if err != nil {
			return fmt.Errorf("soft delete withdrawal wallet address %s: %w", walletAddressID, err)
		}

		err = s.updateProcessingWithdrawalWalletWhitelist(ctx, walletAddress.WithdrawalWalletID, user, totp, tx)
		if err != nil {
			return fmt.Errorf("update withdrawal wallet address %s: %w", walletAddressID, err)
		}
		return nil
	})
}

func (s Service) BatchDelete(ctx context.Context, withdrawalWalletID uuid.UUID, addressesIDs []uuid.UUID, user *models.User, totp string) error {
	return repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		err := s.storage.WithdrawalWalletAddresses(repos.WithTx(tx)).SoftBatchDelete(ctx, addressesIDs)
		if err != nil {
			return fmt.Errorf("soft batch delete withdrawal wallet addresses: %w", err)
		}
		err = s.updateProcessingWithdrawalWalletWhitelist(ctx, withdrawalWalletID, user, totp, tx)
		if err != nil {
			return fmt.Errorf("update withdrawal wallet address %s: %w", withdrawalWalletID, err)
		}
		return nil
	})
}

func (s Service) UpdateAddress(ctx context.Context, params repo_withdrawal_wallet_addresses.UpdateDeletedAddressParams, opts ...repos.Option) (*models.WithdrawalWalletAddress, error) {
	updatedAddress, err := s.storage.WithdrawalWalletAddresses(opts...).UpdateDeletedAddress(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update withdrawal wallet address: %w", err)
	}
	return updatedAddress, nil
}

func (s Service) GetAddressByID(ctx context.Context, id uuid.UUID) (*models.WithdrawalWalletAddress, error) {
	address, err := s.storage.WithdrawalWalletAddresses().GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawal wallet address: %w", err)
	}

	return address, nil
}

func (s Service) GetWalletByID(ctx context.Context, id uuid.UUID) (*models.WithdrawalWallet, error) {
	wallet, err := s.storage.WithdrawalWallets().GetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawal wallet address: %w", err)
	}

	return wallet, nil
}

func (s Service) BatchCreateOrUpdateWallet(ctx context.Context, user *models.User, dto UpdateAddressesListDTO) error {
	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		// create/replace addresses
		if err := s.processWalletAddresses(ctx, dto.Addresses, dto.WalletID, tx); err != nil {
			return fmt.Errorf("process wallet addresses :%w", err)
		}

		// Update processing whitelist
		if err := s.updateProcessingWithdrawalWalletWhitelist(ctx, dto.WalletID, user, dto.TOTP, tx); err != nil {
			return fmt.Errorf("failed to update processing wallet whitelist: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("batch create or update wallet addresses failed: %w", err)
	}

	return nil
}

func (s Service) processWalletAddresses(ctx context.Context, list []AddressDTO, walletID uuid.UUID, tx pgx.Tx) error {
	params := make([]repo_withdrawal_wallet_addresses.UpdateListParams, 0, len(list))
	actualAddresses := make([]string, 0, len(list))
	for _, dto := range list {
		params = append(params, repo_withdrawal_wallet_addresses.UpdateListParams{
			Address:            dto.Address,
			Name:               dto.Name,
			WithdrawalWalletID: walletID,
		})
		actualAddresses = append(actualAddresses, dto.Address)
	}

	deleteParams := repo_withdrawal_wallet_addresses.SoftDeleteUnmatchedByAddressParams{
		WalletID: walletID,
		Address:  actualAddresses,
	}
	if err := s.storage.WithdrawalWalletAddresses(repos.WithTx(tx)).SoftDeleteUnmatchedByAddress(ctx, deleteParams); err != nil {
		return fmt.Errorf("remove unmatched addresses: %w", err)
	}

	errChan := make(chan error, len(params))
	result := s.storage.WithdrawalWalletAddresses(repos.WithTx(tx)).UpdateList(ctx, params)

	wg := sync.WaitGroup{}
	wg.Add(len(params))

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
			return fmt.Errorf("batch create withdrawal wallets: %w", err)
		}
	}

	return nil
}

func (s Service) updateProcessingWithdrawalWalletWhitelist(ctx context.Context, withdrawalWalletID uuid.UUID, user *models.User, totp string, tx pgx.Tx) error {
	wallet, err := s.GetWalletByID(ctx, withdrawalWalletID)
	if err != nil {
		return err
	}
	address, err := s.storage.WithdrawalWalletAddresses(repos.WithTx(tx)).GetWithdrawalWalletsByBlockchain(ctx, repo_withdrawal_wallet_addresses.GetWithdrawalWalletsByBlockchainParams{
		Blockchain: wallet.Blockchain,
		UserID:     wallet.UserID,
	})
	if err != nil {
		return fmt.Errorf("failed to get withdrawal wallet addresses: %w", err)
	}
	params := processing.AttachOwnerColdWalletsParams{
		OwnerID:    user.ProcessingOwnerID.UUID,
		Blockchain: wallet.Blockchain,
		Addresses:  address,
		TOTP:       totp,
	}

	err = s.processingWalletService.AttachOwnerColdWallets(ctx, params)
	if err != nil {
		return fmt.Errorf("update withdrawal wallet processing whitelist: %w", err)
	}
	return nil
}

func (s Service) createWithdrawalWallet(ctx context.Context, c *models.Currency, userID uuid.UUID, opts ...repos.Option) (newWallet *models.WithdrawalWallet, err error) {
	withdrawalMinBalance, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
		Source:     models.RateSourceBinance.String(),
		From:       models.CurrencyCodeUSDT,
		To:         c.Code,
		Amount:     "1",
		StableCoin: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert min balance: %w", err)
	}

	params := repo_withdrawal_wallets.CreateParams{
		UserID:                  userID,
		CurrencyID:              c.ID,
		Blockchain:              *c.Blockchain,
		WithdrawalEnabled:       models.WithdrawalStatusEnabled.String(),
		WithdrawalMinBalance:    decimal.NullDecimal{Decimal: withdrawalMinBalance, Valid: true},
		WithdrawalMinBalanceUsd: decimal.NullDecimal{Decimal: decimal.NewFromInt(1), Valid: true},
		WithdrawalInterval:      models.WithdrawalIntervalNever.String(),
	}

	newWallet, err = s.storage.WithdrawalWallets(opts...).Create(ctx, params)
	if err != nil {
		s.logger.Error("Failed to create withdrawal wallet",
			"user_id", userID,
			"currency_id", c.ID,
			"error", err)
		return nil, fmt.Errorf("failed to create withdrawal wallet for currency %s: %w", c.ID, err)
	}
	return newWallet, nil
}

func (s Service) buildWithdrawalWithAddress(ctx context.Context, wallet *models.WithdrawalWallet, currency *models.CurrencyShort, opts ...repos.Option) (*WithdrawalWithAddress, error) {
	addresses, err := s.storage.WithdrawalWalletAddresses(opts...).GetAddresses(ctx, wallet.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get withdrawal addresses: %w", err)
	}

	multiWithdrawalRules, err := s.storage.MultiWithdrawalRules(opts...).GetByWalletID(ctx, wallet.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get multi withdrawal rules: %w", err)
	}

	multiRuleDTO := MultiWithdrawalRuleDTO{
		Mode: multiWithdrawalRules.Mode,
	}
	if multiWithdrawalRules.ManualAddress.Valid {
		multiRuleDTO.ManualAddress = &multiWithdrawalRules.ManualAddress.String
	}

	rate, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
		Source:     models.RateSourceBinance.String(),
		From:       currency.Code,
		To:         models.CurrencyCodeUSDT,
		Amount:     "1",
		StableCoin: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to convert min balance: %w", err)
	}

	return &WithdrawalWithAddress{
		ID:               wallet.ID,
		Status:           models.WithdrawalStatus(wallet.WithdrawalEnabled),
		MinBalanceNative: wallet.WithdrawalMinBalance,
		MinBalanceUsd:    wallet.WithdrawalMinBalanceUsd,
		Interval:         models.WithdrawalInterval(wallet.WithdrawalInterval),
		Rate:             rate,
		Currency:         *currency,
		Addressees:       addresses,
		MultiWithdrawal:  multiRuleDTO,
	}, nil
}
