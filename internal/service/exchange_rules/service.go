package exchange_rules

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/exchange_manager"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_chains"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchanges"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

type IExchangeRules interface {
	Run(context.Context)
	GetWithdrawalRules(context.Context, models.ExchangeSlug, string) ([]*models.WithdrawalRulesDTO, error)
	GetAllWithdrawalRules(context.Context, uuid.UUID) ([]*models.WithdrawalRulesDTO, error)
	GetWithdrawalRule(context.Context, models.ExchangeSlug, string, string) (*models.WithdrawalRulesDTO, error)
	GetSpotOrderRule(context.Context, models.ExchangeSlug, string) (*models.OrderRulesDTO, error)
}

func NewService(logger logger.Logger, storage storage.IStorage, manager exchange_manager.IExchangeManager) IExchangeRules {
	return &Service{
		logger:  logger,
		storage: storage,
		manager: manager,
	}
}

type Service struct {
	logger  logger.Logger
	storage storage.IStorage
	manager exchange_manager.IExchangeManager
}

func (o *Service) GetAllWithdrawalRules(ctx context.Context, userID uuid.UUID) ([]*models.WithdrawalRulesDTO, error) {
	var dto []*models.WithdrawalRulesDTO
	exchanges, err := o.storage.Exchanges().GetAllActiveWithUserKeys(ctx, userID)
	exchanges = slices.CompactFunc(exchanges, func(lo *repo_exchanges.GetAllActiveWithUserKeysRow, ro *repo_exchanges.GetAllActiveWithUserKeysRow) bool {
		return lo.Slug == ro.Slug
	})
	if err != nil {
		return nil, err
	}
	for _, exchange := range exchanges {
		if exchange.Value.String == "" {
			continue
		}
		enabledCurrencies, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, exchange.Slug)
		if err != nil {
			return nil, err
		}
		rules, err := o.getWithdrawalRules(ctx, exchange.Slug, userID.String(), unwrapCurrencies(enabledCurrencies)...)
		if err != nil {
			return nil, err
		}
		dto = append(dto, rules...)
	}

	return dto, nil
}

func (o *Service) GetWithdrawalRule(ctx context.Context, exchange models.ExchangeSlug, userID string, currency string) (*models.WithdrawalRulesDTO, error) {
	_, err := o.storage.ExchangeChains().GetTickerByCurrencyID(ctx, repo_exchange_chains.GetTickerByCurrencyIDParams{CurrencyID: currency, Slug: exchange})
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("currency %s not found", currency)
	}
	rules, err := o.getWithdrawalRules(ctx, exchange, userID, currency)
	if err != nil || len(rules) == 0 {
		return nil, fmt.Errorf("no rules found for currency %s", currency)
	}
	return rules[0], nil
}

func (o *Service) GetWithdrawalRules(ctx context.Context, exchange models.ExchangeSlug, userID string) ([]*models.WithdrawalRulesDTO, error) {
	enabledCurrencies, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, exchange)
	if err != nil {
		return nil, err
	}
	return o.getWithdrawalRules(ctx, exchange, userID, unwrapCurrencies(enabledCurrencies)...)
}

func (o *Service) Run(ctx context.Context) {
	go func() {
		if err := o.updateHtxDefaultWithdrawalRules(ctx); err != nil {
			o.logger.Error("failed to update default exchange rules", err)
		}
		if err := o.updateUsersWithdrawalRules(ctx); err != nil {
			o.logger.Error("failed to update user exchange rules", err)
		}
	}()
	ticker := time.NewTicker(30 * time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := o.updateHtxDefaultWithdrawalRules(ctx); err != nil {
				o.logger.Error("failed to update default htx exchange rules", err)
			}
			if err := o.updateUsersWithdrawalRules(ctx); err != nil {
				o.logger.Error("failed to update user exchange rules", err)
			}
		}
	}
}
