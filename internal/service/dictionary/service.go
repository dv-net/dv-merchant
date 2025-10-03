package dictionary

import (
	"context"
	"errors"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/internal/service/system"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/pkg/key_value"

	"github.com/redis/go-redis/v9"
)

type IDictionaryService interface {
	LoadDictionary(ctx context.Context) (*Dictionary, error)
}

type Dictionary struct {
	AvailableSources    []string                `json:"available_sources"`
	AvailableCurrencies []*models.CurrencyShort `json:"available_currencies"`
	BackendAddress      string                  `json:"backend_address"`
}

type Service struct {
	storage   storage.IStorage
	rateSvc   exrate.IExRateSource
	systemSvc system.ISystemService
}

var _ IDictionaryService = (*Service)(nil)

func New(storage storage.IStorage, rateSvc exrate.IExRateSource, systemSvc system.ISystemService) IDictionaryService {
	return &Service{
		storage:   storage,
		rateSvc:   rateSvc,
		systemSvc: systemSvc,
	}
}

func (s *Service) LoadDictionary(ctx context.Context) (*Dictionary, error) {
	currencies, err := s.loadAvailableCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	sources, err := s.loadAvailableRateSources(ctx)
	if err != nil {
		return nil, err
	}

	ip, err := s.loadBackendIP(ctx)
	if err != nil {
		return nil, err
	}

	return &Dictionary{
		AvailableSources:    sources,
		AvailableCurrencies: currencies,
		BackendAddress:      ip,
	}, nil
}

func (s *Service) loadBackendIP(ctx context.Context) (string, error) {
	ip, err := s.storage.KeyValue().Get(ctx, "backend_ip")
	if err != nil && (errors.Is(err, redis.Nil) || errors.Is(err, key_value.ErrEntryNotFound)) {
		exIP, err := s.systemSvc.GetIP(ctx)
		if err != nil {
			return "", err
		}
		if err := s.storage.KeyValue().Set(ctx, "backend_ip", exIP, time.Hour*24); err != nil {
			return "", err
		}
		return "", err
	}
	return ip.String(), err
}

func (s *Service) loadAvailableCurrencies(ctx context.Context) ([]*models.CurrencyShort, error) {
	currencies, err := s.storage.Currencies().GetCurrenciesEnabled(ctx)
	if err != nil {
		return nil, err
	}
	shortCurrencies := make([]*models.CurrencyShort, 0, len(currencies))
	for _, currency := range currencies {
		shortCurrencies = append(shortCurrencies, &models.CurrencyShort{
			ID:            currency.ID,
			Code:          currency.Code,
			Precision:     currency.Precision,
			Name:          currency.Name,
			Blockchain:    currency.Blockchain,
			IsBitcoinLike: currency.Blockchain.IsBitcoinLike(),
			IsEVMLike:     currency.Blockchain.IsEVMLike(),
			IsStableCoin:  currency.IsStablecoin,
		})
	}
	return shortCurrencies, nil
}

func (s *Service) loadAvailableRateSources(ctx context.Context) ([]string, error) {
	return s.rateSvc.LoadSources(ctx)
}
