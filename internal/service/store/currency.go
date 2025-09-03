package store

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_currencies"

	"github.com/google/uuid"
)

type IStoreCurrency interface {
	GetCurrenciesByStoreID(ctx context.Context, storeID uuid.UUID) ([]*models.StoreCurrency, error)
	DeleteStoreCurrency(ctx context.Context, store *models.Store, currency *models.Currency, opts ...repos.Option) error
	UpdateStoreCurrency(ctx context.Context, store *models.Store, dto *UpdateStoreCurrencyDTO, opts ...repos.Option) error
	GetAllByStoreID(ctx context.Context, storeID uuid.UUID) ([]*models.Currency, error)
	GetCurrencyWithRate(ctx context.Context, store models.Store, currID string) (*CurrencyRate, error)
}

func (s *Service) GetCurrenciesByStoreID(ctx context.Context, storeID uuid.UUID) ([]*models.StoreCurrency, error) {
	storeCurrencies, err := s.storage.StoreCurrencies().FindAllByStoreID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	return storeCurrencies, nil
}

func (s *Service) CreateStoreCurrency(ctx context.Context, store *models.Store, currency *models.Currency, opts ...repos.Option) error {
	params := repo_store_currencies.CreateOneParams{
		CurrencyID: currency.ID,
		StoreID:    store.ID,
	}
	err := s.storage.StoreCurrencies(opts...).CreateOne(ctx, params)
	if err != nil {
		return fmt.Errorf("update currency: %w", err)
	}

	return nil
}

func (s *Service) DeleteStoreCurrency(ctx context.Context, store *models.Store, currency *models.Currency, opts ...repos.Option) error {
	params := repo_store_currencies.DeleteParams{
		CurrencyID: currency.ID,
		StoreID:    store.ID,
	}
	err := s.storage.StoreCurrencies(opts...).Delete(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) UpdateStoreCurrency(ctx context.Context, store *models.Store, dto *UpdateStoreCurrencyDTO, opts ...repos.Option) error {
	allCurrencies, err := s.currencyService.GetCurrenciesEnabled(ctx)
	if err != nil {
		return err
	}

	enabledCurrencyMap := make(map[string]struct{}, len(dto.CurrencyIDs))
	for _, id := range dto.CurrencyIDs {
		enabledCurrencyMap[id] = struct{}{}
	}
	for _, currency := range allCurrencies {
		_, shouldEnable := enabledCurrencyMap[currency.ID]

		if shouldEnable {
			if err := s.CreateStoreCurrency(ctx, store, currency, opts...); err != nil {
				return err
			}
		} else {
			if err := s.DeleteStoreCurrency(ctx, store, currency, opts...); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetAllByStoreID returns all store currencies by store id
func (s *Service) GetAllByStoreID(ctx context.Context, storeID uuid.UUID) ([]*models.Currency, error) {
	storeCurrencies, err := s.storage.StoreCurrencies().GetAllByStoreID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	return storeCurrencies, nil
}

func (s *Service) GetCurrencyWithRate(ctx context.Context, store models.Store, currID string) (*CurrencyRate, error) {
	curr, err := s.storage.Currencies().GetByID(ctx, currID)
	if err != nil {
		return nil, fmt.Errorf("fetch curr: %w", err)
	}

	rate, err := s.exRate.GetCurrencyRate(ctx, store.RateSource.String(), curr.Code, models.CurrencyCodeUSD)
	if err != nil {
		return nil, fmt.Errorf("fetch rate: %w", err)
	}

	return &CurrencyRate{
		Code:       curr.Code,
		RateSource: store.RateSource.String(),
		Rate:       rate,
	}, nil
}
