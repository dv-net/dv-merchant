package wallet

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallets"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

func (s *Service) GetWallet(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	wallet, err := s.storage.Wallets().GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *Service) UpdateLocale(ctx context.Context, walletID uuid.UUID, locale string, opts ...repos.Option) error {
	return s.storage.Wallets(opts...).UpdateUserLocale(ctx, repo_wallets.UpdateUserLocaleParams{
		Locale: locale,
		ID:     walletID,
	})
}

// GetFullDataByID returns wallet with store, addresses and available currencies
func (s *Service) GetFullDataByID(ctx context.Context, id uuid.UUID) (*GetAllByStoreIDResponse, error) {
	data, err := s.storage.Wallets().GetFullDataByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get full data by id: %w", err)
	}

	availableCurrencies, err := s.storage.StoreCurrencies().GetAllByStoreID(ctx, data.Store.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available currencies by store id: %w", err)
	}

	availableCurrenciesIDs := make([]string, 0, len(availableCurrencies))
	for _, c := range availableCurrencies {
		availableCurrenciesIDs = append(availableCurrenciesIDs, c.ID)
	}

	rates, err := s.exrateService.GetStoreCurrencyRate(ctx, availableCurrencies, data.Store.RateSource.String(), data.Store.RateScale)
	if err != nil {
		return nil, fmt.Errorf("failed to get store currency rate: %w", err)
	}

	addresses, err := s.storage.WalletAddresses().GetAllClearByWalletID(ctx, id, availableCurrenciesIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get all clear addresses by wallet id: %w", err)
	}

	var storeOwner *models.User
	for _, c := range availableCurrencies {
		if c.IsFiat {
			continue
		}

		exists := slices.ContainsFunc(addresses, func(wa *models.WalletAddress) bool {
			return wa.CurrencyID == c.ID
		})
		if exists {
			continue
		}

		if storeOwner == nil {
			storeOwner, err = s.storage.Users().GetByID(ctx, data.Store.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get store owner by id: %w", err)
			}
		}

		newWalletAddress, err := s.getOrCreateWalletAddress(ctx, nil, storeOwner, &data.Wallet, c)
		if err != nil {
			return nil, fmt.Errorf("failed to get or create wallet address: %w", err)
		}

		addresses = append(addresses, newWalletAddress)
	}

	return &GetAllByStoreIDResponse{
		GetFullDataByIDRow:  *data,
		Addresses:           addresses,
		AvailableCurrencies: availableCurrencies,
		Rates:               rates,
	}, nil
}

func (s *Service) GetWalletsInfo(ctx context.Context, usr *models.User, searchCriteria string) ([]*WithBlockchains, error) {
	walletsData, err := s.storage.Wallets().SearchByParam(ctx, repo_wallets.SearchByParamParams{
		Criteria: pgtype.Text{String: searchCriteria, Valid: true},
		UserID:   usr.ID,
	})
	if errors.Is(err, pgx.ErrNoRows) || len(walletsData) == 0 {
		return nil, ErrServiceWalletNotFound
	}
	if err != nil {
		return nil, err
	}

	return s.groupWalletsData(ctx, walletsData, usr.RateSource.String())
}

func (s *Service) groupWalletsData(ctx context.Context, rows []*repo_wallets.SearchByParamRow, rateSource string) ([]*WithBlockchains, error) {
	walletMap := make(map[string]*WithBlockchains)

	for _, row := range rows {
		if _, exists := walletMap[row.Address]; !exists {
			logs, err := s.GetWalletLogs(ctx, row.WalletAddressID)
			if err != nil {
				return nil, fmt.Errorf("error fetching logs for wallet %s: %w", row.WalletID, err)
			}

			w := &WithBlockchains{
				WalletCreatedAt: row.WalletCreatedAt.Time,
				WalletID:        row.WalletID,
				StoreExternalID: row.StoreExternalID,
				StoreID:         row.StoreID,
				StoreName:       row.StoreName,
				Address:         row.Address,
				TotalTx:         decimal.Zero,
				Blockchains:     []BlockchainGroup{},
				Logs:            logs,
			}
			if row.Email.Valid {
				email := row.Email.String
				w.Email = &email
			}
			walletMap[row.Address] = w
		}

		amountUsd, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
			Source:     rateSource,
			From:       row.CurrencyCode,
			To:         models.CurrencyCodeUSDT,
			Amount:     row.Amount.String(),
			StableCoin: false,
		})
		if err != nil {
			return nil, fmt.Errorf("error converting amount for wallet %s: %w", row.WalletID, err)
		}

		w := walletMap[row.Address]
		asset := AssetWallet{
			Currency:     row.CurrencyID,
			Amount:       row.Amount,
			AmountUSD:    amountUsd,
			TxCount:      row.DepositsCount,
			TotalDeposit: row.DepositsSum,
		}

		found := false
		for i, bg := range w.Blockchains {
			if bg.Blockchain == row.Blockchain {
				w.Blockchains[i].Assets = append(w.Blockchains[i].Assets, asset)
				found = true
				break
			}
		}
		if !found {
			w.Blockchains = append(w.Blockchains, BlockchainGroup{
				Blockchain: row.Blockchain,
				Assets:     []AssetWallet{asset},
			})
		}
	}

	result := make([]*WithBlockchains, 0, len(walletMap))
	for _, w := range walletMap {
		w.TotalTx = lo.Reduce(w.Blockchains, func(acc decimal.Decimal, bg BlockchainGroup, _ int) decimal.Decimal {
			return acc.Add(lo.Reduce(bg.Assets, func(acc decimal.Decimal, asset AssetWallet, _ int) decimal.Decimal {
				return acc.Add(asset.TxCount)
			}, decimal.Zero))
		}, decimal.Zero)
		result = append(result, w)
	}

	return result, nil
}
