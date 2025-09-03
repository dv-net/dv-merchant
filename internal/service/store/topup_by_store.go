package store

import (
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/wallet"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/net/context"
)

type TopUpData struct {
	Store               *models.Store
	Addresses           []*models.WalletAddress
	AvailableCurrencies []*models.Currency
	Rates               map[string]string
	WalletID            uuid.UUID
}

var (
	ErrStoreRequestLimitExceeded = errors.New("rate limit exceeded")
	ErrStoreNotFound             = errors.New("store not found")
	ErrStoreDisabled             = errors.New("store is disabled")
)

func (s *Service) PrepareTopUpDataByStore(ctx context.Context, dto wallet.CreateStoreWalletWithAddressDTO) (*TopUpData, error) {
	store, err := s.storage.Stores().GetByIDWithPublicFormEnabled(ctx, dto.StoreID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrStoreNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("fetch store: %w", err)
	}

	if !store.Status {
		return nil, ErrStoreDisabled
	}

	if s.rateLimitEnabled && !s.rateLimiter.IsAllowedByKey(ctx, fmt.Sprintf("store_rate_limit_%s", store.ID.String())) {
		return nil, ErrStoreRequestLimitExceeded
	}

	res, err := s.wallets.StoreWalletWithAddress(ctx, dto, "0")
	if err != nil {
		return nil, fmt.Errorf("prepare store wallts: %w", err)
	}

	availableCurrencies, err := s.GetAllByStoreID(ctx, store.ID)
	if err != nil {
		return nil, fmt.Errorf("fetch store data: %w", err)
	}

	// get rates data
	var rates map[string]string
	rates, err = s.exRate.GetStoreCurrencyRate(ctx, availableCurrencies, store.RateSource.String(), store.RateScale)
	if err != nil {
		return nil, fmt.Errorf("fetch store data: %w", err)
	}

	return &TopUpData{
		Store:               store,
		WalletID:            res.ID,
		Addresses:           res.Address,
		AvailableCurrencies: availableCurrencies,
		Rates:               rates,
	}, nil
}
