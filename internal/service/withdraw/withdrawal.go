package withdraw

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_withdrawal_from_processing_wallets"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_withdrawal_wallet_addresses"

	"connectrpc.com/connect"
	"github.com/dv-net/dv-processing/pkg/avalidator"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
)

type IWithdrawalService interface {
	WithdrawFromAddress(ctx context.Context, user *models.User, walletID uuid.UUID, currencyID string) error
	WithdrawFromAddresses(ctx context.Context, user *models.User, dto MultipleWithdrawalDTO) error
	WithdrawToProcessingWallet(ctx context.Context, user *models.User, dto WithdrawalToProcessingDTO) error
	CreateWithdrawalFromProcessing(ctx context.Context, dto CreateWithdrawalFromProcessingDTO) (*models.WithdrawalFromProcessingWallet, error)
	DeleteWithdrawalFromProcessing(ctx context.Context, id uuid.UUID, storeID uuid.UUID) error
	GetProcessingWithdrawalWithTransfer(ctx context.Context, requestID string, storeID uuid.UUID) (*WithdrawalFromProcessingDto, error)
}

func (s *service) WithdrawFromAddress(
	ctx context.Context,
	user *models.User,
	walletAddressID uuid.UUID,
	currencyID string,
) error {
	curr, err := s.currencyService.GetCurrencyByID(ctx, currencyID)
	if err != nil {
		return fmt.Errorf("fetch currency: %w", err)
	}
	if curr.Blockchain == nil {
		return ErrFiatCurrencyIsNotSupported
	}

	targetWalletAddress, err := s.storage.WalletAddresses().GetById(ctx, walletAddressID)
	if err != nil {
		return fmt.Errorf("fetch wallet: %w", err)
	}

	if targetWalletAddress.UserID != user.ID {
		return ErrWalletIsNotOwnedByUser
	}

	wallets, err := s.storage.WithdrawalWalletAddresses().GetWithdrawalAddressByCurrencyID(
		ctx,
		repo_withdrawal_wallet_addresses.GetWithdrawalAddressByCurrencyIDParams{
			Blockchain: targetWalletAddress.Blockchain,
			UserID:     user.ID,
			CurrencyID: currencyID,
		},
	)
	if err != nil || len(wallets) < 1 {
		return fmt.Errorf("wallet for withdrawal not found")
	}

	decRate, err := s.currencyRate(ctx, user.RateSource.String(), curr)
	if err != nil {
		return err
	}

	dto := TransferDto{
		ID:            uuid.New(),
		UserID:        user.ID,
		OwnerID:       user.ProcessingOwnerID.UUID,
		Kind:          models.TransferKindFromAddress,
		FromAddresses: []string{targetWalletAddress.Address},
		ToAddress:     s.prepareWithdrawalAddressByUsed(ctx, targetWalletAddress.Address, wallets),
		Contract:      curr.ContractAddress.String,
		Amount:        targetWalletAddress.Amount,
		AmountUsd:     decRate.Mul(targetWalletAddress.Amount),
		CurrencyID:    curr.ID,
		Blockchain:    targetWalletAddress.Blockchain,
	}

	transfer, err := s.initializeTransfer(ctx, dto, user, nil)
	if err != nil {
		return fmt.Errorf("transfer init: %w", err)
	}

	if transfer.Status == models.TransferStatusFailed {
		var msg string
		if transfer.Message != nil {
			msg = *transfer.Message
		}

		return fmt.Errorf("withdrawal failed: %s", msg)
	}

	return nil
}

func (s *service) WithdrawFromAddresses(ctx context.Context, user *models.User, dto MultipleWithdrawalDTO) error {
	params := repo_wallet_addresses.GetListByCurrencyWithAmountParams{
		CurrID:      dto.CurrencyID,
		UserID:      user.ID,
		ExcludedIds: dto.ExcludedWalletAddressesIDs,
		Ids:         dto.WalletAddressIDs,
	}

	walletList, err := s.storage.WalletAddresses().GetListByCurrencyWithAmount(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrWithdrawalAddressEmptyBalances
		}
		return err
	}

	if len(dto.WalletAddressIDs) > 0 && len(walletList.Addresses) != len(dto.WalletAddressIDs) {
		return errors.New("invalid addresses")
	}

	if len(dto.WalletAddressIDs) > 0 && !walletList.Currency.Blockchain.IsBitcoinLike() {
		return ErrTransferFromMultipleAddressNotSupported
	}

	decRate, err := s.currencyRate(ctx, user.RateSource.String(), &walletList.Currency)
	if err != nil {
		return err
	}

	withdrawalAddr, err := s.storage.WithdrawalWalletAddresses().GetById(ctx, dto.WithdrawalWalletID)
	if err != nil {
		return err
	}

	transferDto := TransferDto{
		ID:            uuid.New(),
		UserID:        user.ID,
		OwnerID:       user.ProcessingOwnerID.UUID,
		Kind:          models.TransferKindFromAddress,
		FromAddresses: walletList.Addresses,
		ToAddress:     withdrawalAddr.Address,
		Contract:      walletList.Currency.ContractAddress.String,
		Amount:        walletList.Amount,
		AmountUsd:     decRate.Mul(walletList.Amount),
		CurrencyID:    walletList.Currency.ID,
		Blockchain:    *walletList.Currency.Blockchain,
	}

	transfer, err := s.initializeTransfer(ctx, transferDto, user, nil)
	if err != nil {
		return fmt.Errorf("transfer init: %w", err)
	}

	if transfer.Status == models.TransferStatusFailed {
		var msg string
		if transfer.Message != nil {
			msg = *transfer.Message
		}

		return fmt.Errorf("withdrawal failed: %s", msg)
	}

	return nil
}

func (s *service) WithdrawToProcessingWallet(ctx context.Context, user *models.User, dto WithdrawalToProcessingDTO) error {
	walletsList, err := s.storage.WalletAddresses().GetListByCurrencyWithAmount(ctx, repo_wallet_addresses.GetListByCurrencyWithAmountParams{
		UserID:      user.ID,
		CurrID:      dto.CurrencyID,
		Ids:         dto.WalletAddressIDs,
		ExcludedIds: dto.ExcludedWalletAddressesIDs,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrWithdrawalAddressEmptyBalances
		}
		return fmt.Errorf("fetch wallet: %w", err)
	}

	if walletsList.Currency.Blockchain == nil {
		return ErrFiatCurrencyIsNotSupported
	}

	if len(dto.WalletAddressIDs) > 1 && !walletsList.Currency.Blockchain.IsBitcoinLike() {
		return ErrTransferFromMultipleAddressNotSupported
	}

	processingWallet, processingErr := s.processingWallet.GetOwnerProcessingWallet(ctx, processing.GetOwnerProcessingWalletsParams{
		OwnerID:    user.ProcessingOwnerID.UUID,
		Blockchain: walletsList.Currency.Blockchain,
	})
	if processingErr != nil {
		if connect.CodeOf(processingErr) == connect.CodeUnavailable {
			return ErrProcessingUnavailable
		}
		return fmt.Errorf("fetch wallet from processing: %w", processingErr)
	}

	decRate, err := s.currencyRate(ctx, user.RateSource.String(), &walletsList.Currency)
	if err != nil {
		return fmt.Errorf("fetch currency rate: %w", err)
	}

	transferDTO := TransferDto{
		ID:            uuid.New(),
		UserID:        user.ID,
		OwnerID:       user.ProcessingOwnerID.UUID,
		Kind:          models.TransferKindFromAddress,
		FromAddresses: walletsList.Addresses,
		ToAddress:     processingWallet.Address,
		Contract:      walletsList.Currency.ContractAddress.String,
		Amount:        walletsList.Amount,
		AmountUsd:     decRate.Mul(walletsList.Amount),
		CurrencyID:    walletsList.Currency.ID,
		Blockchain:    *walletsList.Currency.Blockchain,
	}

	transfer, err := s.initializeTransfer(ctx, transferDTO, user, nil)
	if err != nil {
		return fmt.Errorf("transfer init: %w", err)
	}

	if transfer.Status == models.TransferStatusFailed {
		var msg string
		if transfer.Message != nil {
			msg = *transfer.Message
		}

		return fmt.Errorf("withdrawal failed: %s", msg)
	}

	return nil
}

func (s *service) CreateWithdrawalFromProcessing(ctx context.Context, dto CreateWithdrawalFromProcessingDTO) (*models.WithdrawalFromProcessingWallet, error) {
	usr, err := s.storage.Users().GetByID(ctx, dto.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetch user: %w", err)
	}

	stores, err := s.storage.Stores().GetByUser(ctx, dto.UserID)
	if err != nil {
		return nil, fmt.Errorf("fetch stores: %w", err)
	}

	if dto.StoreID != nil {
		if !lo.ContainsBy(stores, func(s *models.Store) bool {
			return s.ID == *dto.StoreID && s.UserID == dto.UserID
		}) {
			return nil, ErrStoreIsNotOwnedByUser
		}
	}

	if !lo.ContainsBy(stores, func(s *models.Store) bool {
		return s.UserID == dto.UserID
	}) {
		return nil, ErrStoreIsNotOwnedByUser
	}

	// Check is processing withdrawals enabled by owner
	res, err := s.settings.GetModelSetting(ctx, setting.WithdrawFromProcessing, setting.IModelSetting(usr))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("API withdrawal unavailable: this feature is disabled in the application settings")
	}
	if err != nil {
		return nil, fmt.Errorf("get setting: %w", err)
	}
	if res == nil || res.Value == setting.FlagValueDisabled {
		return nil, ErrWithdrawalsFromProcessingDisabled
	}

	curr, err := s.currencyService.GetCurrencyByID(ctx, dto.CurrencyID)
	if err != nil {
		return nil, fmt.Errorf("fetch currency: %w", err)
	}

	exist, err := s.storage.WithdrawalsFromProcessing().IsWithdrawalExistByRequestID(ctx, dto.RequestID)
	if err != nil {
		return nil, fmt.Errorf("check if request id exist: %w", err)
	}
	if exist {
		return nil, ErrWithdrawFromProcessingDuplicateRequestID
	}

	// Validate address is valid by blockchain
	if !avalidator.ValidateAddressByBlockchain(dto.AddressTo, curr.Blockchain.String()) {
		return nil, &InvalidCurrencyForAddressError{
			Wallet:     dto.AddressTo,
			Blockchain: curr.Blockchain.String(),
		}
	}

	hotWalletExists, err := s.storage.WalletAddresses().IsWalletExistsByAddress(ctx, dto.AddressTo)
	if err != nil {
		return nil, fmt.Errorf("check wallet exists: %w", err)
	}
	if hotWalletExists {
		return nil, ErrWithdrawFromProcessingToHotNotAllowed
	}

	if !usr.ProcessingOwnerID.Valid {
		return nil, ErrProcessingUninitialized
	}

	targetWallets, err := s.processingWallet.GetOwnerProcessingWallets(ctx, processing.GetOwnerProcessingWalletsParams{
		OwnerID:    usr.ProcessingOwnerID.UUID,
		Blockchain: curr.Blockchain,
	})
	if err != nil {
		return nil, fmt.Errorf("fetch wallets from processing: %w", err)
	}

	if len(targetWallets) == 0 {
		return nil, ErrProcessingWalletNotExists
	}

	wallet := targetWallets[0]
	createParams := repo_withdrawal_from_processing_wallets.CreateParams{
		CurrencyID:  curr.ID,
		AddressFrom: wallet.Address,
		AddressTo:   dto.AddressTo,
		Amount:      dto.Amount,
		RequestID:   dto.RequestID,
	}
	if len(stores) > 0 {
		createParams.StoreID = stores[0].ID
	}
	if dto.StoreID != nil {
		createParams.StoreID = *dto.StoreID
	}
	withdrawal, err := s.storage.WithdrawalsFromProcessing().Create(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("withdrawal creation: %w", err)
	}

	return withdrawal, nil
}

func (s *service) DeleteWithdrawalFromProcessing(ctx context.Context, id uuid.UUID, storeID uuid.UUID) error {
	wfpw, err := s.storage.WithdrawalsFromProcessing().DeleteWithdrawalFromProcessingWallets(ctx, repo_withdrawal_from_processing_wallets.DeleteWithdrawalFromProcessingWalletsParams{
		ID:      id,
		StoreID: storeID,
	})
	// success delete
	if err == nil && wfpw != nil {
		return nil
	}

	record, err := s.storage.WithdrawalsFromProcessing().GetByID(ctx, repo_withdrawal_from_processing_wallets.GetByIDParams{
		ID:      id,
		StoreID: storeID,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrWithdrawalNotFound
		}
		return fmt.Errorf("get record: %w", err)
	}

	if record.TransferID.Valid {
		return ErrWithdrawalCannotBeDeleted
	}

	return fmt.Errorf("undelete withdrawal")
}

func (s *service) GetProcessingWithdrawalWithTransfer(
	ctx context.Context,
	requestID string,
	storeID uuid.UUID,
) (*WithdrawalFromProcessingDto, error) {
	res, err := s.storage.WithdrawalsFromProcessing().GetWithdrawalWithTransfer(
		ctx,
		repo_withdrawal_from_processing_wallets.GetWithdrawalWithTransferParams{
			RequestID: requestID,
			StoreID:   storeID,
		},
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("fetch withdrawal: %w", err)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrWithdrawalNotFound
	}

	var preparedTransferInfo *ShortTransferDto
	if res.WithdrawalFromProcessingWallet.TransferID.Valid {
		preparedTransferInfo = &ShortTransferDto{
			Kind:    res.Kind,
			Status:  res.Status,
			Stage:   res.Stage,
			Message: res.Message,
		}
	}

	var createdAt *time.Time
	if res.WithdrawalFromProcessingWallet.CreatedAt.Valid {
		createdAt = &res.WithdrawalFromProcessingWallet.CreatedAt.Time
	}

	return &WithdrawalFromProcessingDto{
		Transfer:    preparedTransferInfo,
		TXHash:      *res.TxHash,
		StoreID:     res.WithdrawalFromProcessingWallet.StoreID,
		CurrencyID:  res.WithdrawalFromProcessingWallet.CurrencyID,
		AddressFrom: res.WithdrawalFromProcessingWallet.AddressFrom,
		AddressTo:   res.WithdrawalFromProcessingWallet.AddressTo,
		Amount:      res.WithdrawalFromProcessingWallet.Amount,
		AmountUSD:   res.WithdrawalFromProcessingWallet.AmountUsd,
		CreatedAt:   createdAt,
	}, nil
}
