package withdraw

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_withdrawal_from_processing_wallets"
	"github.com/dv-net/dv-merchant/internal/tools"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_transactions"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *service) GetPrefetchWithdrawalAddress(ctx context.Context, user *models.User) ([]*models.PrefetchWithdrawAddressInfo, error) {
	// no prefetch data if transfers disabled by settings
	transferStatusSetting, err := s.settings.GetModelSetting(ctx, setting.TransfersStatus, setting.IModelSetting(user))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("fetch setting: %w", err)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if transferStatusSetting.Value != setting.FlagValueEnabled {
		return nil, nil
	}

	rates, err := s.exRateService.LoadRatesList(ctx, user.RateSource.String())
	if err != nil {
		return nil, fmt.Errorf("failed to load rate: %w", err)
	}

	prefetchData, err := s.getPrefetchData(ctx, user.ID, rates)
	if err != nil {
		return nil, err
	}

	var prefetchProcessingWithdrawals []*repo_withdrawal_from_processing_wallets.GetPrefetchHistoryByUserIDRow
	settingWithdrawProcessing, err := s.settings.GetModelSetting(ctx, setting.WithdrawFromProcessing, setting.IModelSetting(user))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("fetch setting: %w", err)
	}
	if err == nil && settingWithdrawProcessing.Value == setting.FlagValueEnabled {
		prefetchProcessingWithdrawals, err = s.storage.WithdrawalsFromProcessing().GetPrefetchHistoryByUserID(
			ctx,
			repo_withdrawal_from_processing_wallets.GetPrefetchHistoryByUserIDParams{
				UserID:       user.ID,
				CurrencyIds:  rates.CurrencyIDs,
				CurrencyRate: rates.Rate,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	multiWithdrawalWallets, err := s.storage.WithdrawalWallets().GetForMultiWithdrawal(ctx, uuid.NullUUID{UUID: user.ID, Valid: true})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("fetch multi withdrawals: %w", err)
	}

	multiWithdrawalPrefetch := make([]*models.PrefetchWithdrawAddressInfo, 0, len(multiWithdrawalWallets))
	for _, multiWithdrawalWallet := range multiWithdrawalWallets {
		addrs, multiAddrErr := s.storage.WalletAddresses().GetAddressForMultiWithdrawal(ctx, repo_wallet_addresses.GetAddressForMultiWithdrawalParams{
			CurrencyIds:  rates.CurrencyIDs,
			CurrencyRate: rates.Rate,
			UserID:       user.ID,
			Currency:     multiWithdrawalWallet.Currency.ID,
			Blockchain:   *multiWithdrawalWallet.Currency.Blockchain,
			MinUsd:       multiWithdrawalWallet.WithdrawalWallet.WithdrawalMinBalanceUsd,
			MinAmount:    multiWithdrawalWallet.WithdrawalWallet.WithdrawalMinBalance,
		})
		if multiAddrErr != nil {
			if !errors.Is(multiAddrErr, pgx.ErrNoRows) {
				s.logger.Warn("fetch wallet addresses for multi withdrawals failed", multiAddrErr)
			}

			continue
		}

		addr, err := s.prepareWalletToByMultipleRules(ctx, &multiWithdrawalWallet.User, multiWithdrawalWallet.MultiWithdrawalRule, multiWithdrawalWallet.Currency, multiWithdrawalWallet.Addresses)
		if err != nil || addr == "" {
			s.logger.Debug("fetch wallet addresses for multi withdrawals failed", "err", multiAddrErr)
			continue
		}

		multiWithdrawalPrefetch = append(multiWithdrawalPrefetch, &models.PrefetchWithdrawAddressInfo{
			Currency: models.CurrencyShort{
				ID:            multiWithdrawalWallet.Currency.ID,
				Code:          multiWithdrawalWallet.Currency.Code,
				Precision:     multiWithdrawalWallet.Currency.Precision,
				Name:          multiWithdrawalWallet.Currency.Name,
				Blockchain:    multiWithdrawalWallet.Currency.Blockchain,
				IsBitcoinLike: multiWithdrawalWallet.Currency.Blockchain.IsBitcoinLike(),
				IsEVMLike:     multiWithdrawalWallet.Currency.Blockchain.IsEVMLike(),
				IsStableCoin:  multiWithdrawalWallet.Currency.IsStablecoin,
			},
			Amount:      addrs.TotalAmount,
			AmountUsd:   addrs.AmountUsd,
			Type:        models.TransferKindFromAddress,
			AddressFrom: addrs.Addresses,
			AddressTo:   []string{addr},
		})
	}

	data := make([]*models.PrefetchWithdrawAddressInfo, 0, len(prefetchData)+len(prefetchProcessingWithdrawals)+len(multiWithdrawalPrefetch))
	data = append(data, multiWithdrawalPrefetch...)
	for _, prefetchedRow := range prefetchData {
		addressesTo := make([]string, 0, 1)
		if prefetchedRow.WithdrawalWalletID.Valid {
			addr, fetchAddrErr := s.getWithdrawalAddress(ctx, prefetchedRow.WithdrawalWalletID.UUID, prefetchedRow.WalletAddress.Address)
			if fetchAddrErr != nil && !errors.Is(fetchAddrErr, pgx.ErrNoRows) {
				return nil, fmt.Errorf("fetch address for withdrawal: %w", err)
			}

			if addr != nil && *addr != "" {
				addressesTo = append(addressesTo, *addr)
			}
		}

		data = append(data, &models.PrefetchWithdrawAddressInfo{
			Currency: models.CurrencyShort{
				ID:            prefetchedRow.Currency.ID,
				Name:          prefetchedRow.Currency.Name,
				Code:          prefetchedRow.Currency.Code,
				Precision:     prefetchedRow.Currency.Precision,
				Blockchain:    &prefetchedRow.WalletAddress.Blockchain,
				IsBitcoinLike: prefetchedRow.Currency.Blockchain.IsBitcoinLike(),
				IsEVMLike:     prefetchedRow.Currency.Blockchain.IsEVMLike(),
			},
			Amount:      prefetchedRow.WalletAddress.Amount,
			AmountUsd:   prefetchedRow.AmountUsd,
			Type:        models.TransferKindFromAddress,
			AddressFrom: []string{prefetchedRow.WalletAddress.Address},
			AddressTo:   addressesTo,
		})
	}

	for _, processingWithdrawal := range prefetchProcessingWithdrawals {
		data = append(data, &models.PrefetchWithdrawAddressInfo{
			Currency: models.CurrencyShort{
				ID:            processingWithdrawal.Currency.ID,
				Code:          processingWithdrawal.Currency.Code,
				Precision:     processingWithdrawal.Currency.Precision,
				Name:          processingWithdrawal.Currency.Name,
				Blockchain:    processingWithdrawal.Currency.Blockchain,
				IsBitcoinLike: processingWithdrawal.Currency.Blockchain.IsBitcoinLike(),
				IsEVMLike:     processingWithdrawal.Currency.Blockchain.IsEVMLike(),
			},
			Amount:      processingWithdrawal.WithdrawalFromProcessingWallet.Amount,
			AmountUsd:   processingWithdrawal.AmountUsd,
			Type:        models.TransferKindFromProcessing,
			AddressFrom: []string{processingWithdrawal.WithdrawalFromProcessingWallet.AddressFrom},
			AddressTo:   []string{processingWithdrawal.WithdrawalFromProcessingWallet.AddressTo},
		})
	}

	return data, nil
}

func (s *service) prepareWalletToByMultipleRules(
	ctx context.Context,
	u *models.User,
	rule models.MultiWithdrawalRule,
	curr models.Currency,
	withdrawalAddresses []string,
) (string, error) {
	switch rule.Mode {
	case models.MultiWithdrawalModeManual:
		if !rule.ManualAddress.Valid {
			return "", errors.New("manual address for multiple withdrawal is not set")
		}

		return rule.ManualAddress.String, nil
	case models.MultiWithdrawalModeProcessing:
		res, err := s.processingWallet.GetOwnerProcessingWallet(ctx, processing.GetOwnerProcessingWalletsParams{
			OwnerID:    u.ProcessingOwnerID.UUID,
			Blockchain: curr.Blockchain,
		})
		if err != nil {
			return "", err
		}

		return res.Address, nil
	case models.MultiWithdrawalModeRandom:
		if len(withdrawalAddresses) == 0 {
			return "", errors.New("withdrawal addresses list is empty")
		}
		return tools.RandomSliceElement(withdrawalAddresses), nil
	default:
		return "", fmt.Errorf("mode '%s' is not supported", rule.Mode)
	}
}

func (s *service) getPrefetchData(ctx context.Context, userID uuid.UUID, rates *exrate.Rates) ([]*repo_wallet_addresses.GetPrefetchWalletAddressByUserIDRow, error) {
	params := repo_wallet_addresses.GetPrefetchWalletAddressByUserIDParams{
		UserID:       userID,
		CurrencyIds:  rates.CurrencyIDs,
		CurrencyRate: rates.Rate,
	}

	prefetchData, err := s.storage.WalletAddresses().GetPrefetchWalletAddressByUserID(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("get prefetch wallet address: %w", err)
	}

	return prefetchData, nil
}

func (s *service) getWithdrawalAddress(ctx context.Context, withdrawalWalletID uuid.UUID, fromAddr string) (*string, error) {
	withdrawalAddrList, err := s.getWithdrawalAddressList(ctx, withdrawalWalletID)
	if err != nil {
		return nil, err
	}

	addr, err := s.storage.Transactions().GetExistingWithdrawalAddress(
		ctx,
		repo_transactions.GetExistingWithdrawalAddressParams{
			FromAddr:          fromAddr,
			WithdrawAddresses: withdrawalAddrList,
		},
	)
	if err != nil {
		return nil, err
	}

	return &addr, nil
}

func (s *service) getWithdrawalAddressList(ctx context.Context, walletID uuid.UUID) ([]string, error) {
	withdrawalAddrList, err := s.storage.WithdrawalWalletAddresses().GetAddressesList(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("fetch addresses by withdrawal wallet: %w", err)
	}

	return withdrawalAddrList, nil
}
