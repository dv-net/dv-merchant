package currency

import (
	"context"
	"errors"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_currencies"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ICurrency interface {
	GetAllCurrency(ctx context.Context) ([]*models.Currency, error)
	GetCurrencyByID(ctx context.Context, ID string) (*models.Currency, error)
	GetEnableCurrencyByAddress(ctx context.Context, ID string) (*models.Currency, error)
	GetCurrenciesHasBalance(ctx context.Context) ([]*models.Currency, error)
	GetCurrenciesEnabled(ctx context.Context) ([]*models.Currency, error)
	GetCurrencyByBlockchainAndContract(ctx context.Context, blockchain models.Blockchain, contract string) (*models.Currency, error)
	GetCurrenciesByBlockchain(ctx context.Context, blockchain models.Blockchain) ([]*models.Currency, error)
	GetEnabledCurrencyByCode(ctx context.Context, code string, blockchain models.Blockchain) (*models.Currency, error)
}

type Service struct {
	cfg     *config.Config
	storage storage.IStorage
}

type CreateParams = repo_currencies.CreateParams

func New(cfg *config.Config, storage storage.IStorage) *Service {
	return &Service{
		cfg:     cfg,
		storage: storage,
	}
}

func (s Service) GetAllCurrency(ctx context.Context) ([]*models.Currency, error) {
	all, err := s.storage.Currencies().GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return all, nil
}

func (s Service) GetCurrencyByID(ctx context.Context, id string) (*models.Currency, error) {
	currency, err := s.storage.Currencies().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return currency, nil
}

func (s Service) GetEnableCurrencyByAddress(ctx context.Context, id string) (*models.Currency, error) {
	currency, err := s.storage.Currencies().GetEnabledCurrencyById(ctx, id)
	if err != nil {
		return nil, err
	}
	return currency, nil
}

func (s Service) GetCurrenciesHasBalance(ctx context.Context) ([]*models.Currency, error) {
	currencies, err := s.storage.Currencies().GetCurrenciesHasBalance(ctx)
	if err != nil {
		return nil, err
	}
	return currencies, nil
}

func (s Service) GetCurrencyByBlockchainAndContract(ctx context.Context, blockchain models.Blockchain, contract string) (*models.Currency, error) {
	params := repo_currencies.GetCurrencyByBlockchainAndContractParams{
		Blockchain:      &blockchain,
		ContractAddress: pgtype.Text{String: contract, Valid: true},
	}
	currencies, err := s.storage.Currencies().GetCurrencyByBlockchainAndContract(ctx, params)
	if err != nil {
		return nil, err
	}
	return currencies, nil
}

func (s Service) GetCurrenciesEnabled(ctx context.Context) ([]*models.Currency, error) {
	currencies, err := s.storage.Currencies().GetCurrenciesEnabled(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = fmt.Errorf("no currencies enabled: %w", err)
		}
		return nil, err
	}
	return currencies, nil
}

func (s Service) GetCurrenciesByBlockchain(ctx context.Context, blockchain models.Blockchain) ([]*models.Currency, error) {
	currencies, err := s.storage.Currencies().GetCurrenciesByBlockchain(ctx, &blockchain)
	if err != nil {
		return nil, err
	}
	return currencies, nil
}

func (s Service) GetEnabledCurrencyByCode(ctx context.Context, code string, blockchain models.Blockchain) (*models.Currency, error) {
	args := repo_currencies.GetEnabledCurrencyByCodeParams{
		Code:       code,
		Blockchain: util.Pointer(blockchain),
	}

	currency, err := s.storage.Currencies().GetEnabledCurrencyByCode(ctx, args)
	if err != nil {
		return nil, err
	}
	return currency, nil
}
