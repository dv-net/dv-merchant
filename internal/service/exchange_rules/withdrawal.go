package exchange_rules

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/fatih/structs"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_chains"
	"github.com/dv-net/dv-merchant/pkg/key_value"
)

func (o *Service) getWithdrawalRules(ctx context.Context, exchange models.ExchangeSlug, userID string, currencyIDs ...string) ([]*models.WithdrawalRulesDTO, error) {
	rules := make([]*models.WithdrawalRulesDTO, 0, len(currencyIDs))
	for _, currency := range currencyIDs {
		rule, err := o.fetchWithdrawalRule(ctx, exchange, userID, currency)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (o *Service) fetchWithdrawalRule(ctx context.Context, exchange models.ExchangeSlug, userID, currency string) (*models.WithdrawalRulesDTO, error) {
	switch exchange {
	case models.ExchangeSlugHtx:
		return o.handleHtxWithdrawal(ctx, currency)
	case models.ExchangeSlugOkx:
		return o.handleOkxWithdrawal(ctx, userID, currency)
	case models.ExchangeSlugBinance:
		return o.handleBinanceWithdrawal(ctx, userID, currency)
	case models.ExchangeSlugBitget:
		return o.handleBitgetWithdrawal(ctx, userID, currency)
	case models.ExchangeSlugKucoin:
		return o.handleKucoinWithdrawal(ctx, userID, currency)
	case models.ExchangeSlugGateio:
		return o.handleGateioWithdrawal(ctx, userID, currency)
	case models.ExchangeSlugBybit:
		return o.handleBybitWithdrawal(ctx, userID, currency)
	default:
		return nil, fmt.Errorf("exchange %s not supported", exchange.String())
	}
}

func (o *Service) handleHtxWithdrawal(ctx context.Context, currency string) (*models.WithdrawalRulesDTO, error) {
	rule := &models.WithdrawalRulesDTO{}
	data, err := o.storage.KeyValue().Get(ctx, formatKey(models.ExchangeSlugHtx, WithdrawalRuleType, currency, "default"))
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, key_value.ErrEntryNotFound) {
			return o.fetchAndCacheHtxWithdrawalRules(ctx, currency)
		}
		return nil, err
	}
	return rule, json.Unmarshal(data, rule)
}

func (o *Service) handleOkxWithdrawal(ctx context.Context, userID, currency string) (*models.WithdrawalRulesDTO, error) {
	rule := &models.WithdrawalRulesDTO{}
	data, err := o.storage.KeyValue().Get(ctx, formatKey(models.ExchangeSlugOkx, WithdrawalRuleType, currency, userID))
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, key_value.ErrEntryNotFound) {
			return o.fetchAndCacheOkxWithdrawalRules(ctx, userID, currency)
		}
		return nil, err
	}
	return rule, json.Unmarshal(data, rule)
}

func (o *Service) handleBinanceWithdrawal(ctx context.Context, userID, currency string) (*models.WithdrawalRulesDTO, error) {
	rule := &models.WithdrawalRulesDTO{}
	data, err := o.storage.KeyValue().Get(ctx, formatKey(models.ExchangeSlugBinance, WithdrawalRuleType, currency, userID))
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, key_value.ErrEntryNotFound) {
			return o.fetchAndCacheBinanceWithdrawalRules(ctx, userID, currency)
		}
		return nil, err
	}
	return rule, json.Unmarshal(data, rule)
}

func (o *Service) handleBitgetWithdrawal(ctx context.Context, userID, currency string) (*models.WithdrawalRulesDTO, error) {
	rule := &models.WithdrawalRulesDTO{}
	data, err := o.storage.KeyValue().Get(ctx, formatKey(models.ExchangeSlugBitget, WithdrawalRuleType, currency, userID))
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, key_value.ErrEntryNotFound) {
			return o.fetchAndCacheBitgetWithdrawalRules(ctx, userID, currency)
		}
		return nil, err
	}
	return rule, json.Unmarshal(data, rule)
}

func (o *Service) handleKucoinWithdrawal(ctx context.Context, userID, currency string) (*models.WithdrawalRulesDTO, error) {
	rule := &models.WithdrawalRulesDTO{}
	data, err := o.storage.KeyValue().Get(ctx, formatKey(models.ExchangeSlugKucoin, WithdrawalRuleType, currency, userID))
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, key_value.ErrEntryNotFound) {
			return o.fetchAndCacheKucoinWithdrawalRules(ctx, userID, currency)
		}
		return nil, err
	}
	return rule, json.Unmarshal(data, rule)
}

func (o *Service) handleGateioWithdrawal(ctx context.Context, userID, currency string) (*models.WithdrawalRulesDTO, error) {
	rule := &models.WithdrawalRulesDTO{}
	data, err := o.storage.KeyValue().Get(ctx, formatKey(models.ExchangeSlugGateio, WithdrawalRuleType, currency, userID))
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, key_value.ErrEntryNotFound) {
			return o.fetchAndCacheGateioWithdrawalRules(ctx, userID, currency)
		}
		return nil, err
	}
	return rule, json.Unmarshal(data, rule)
}

func (o *Service) handleBybitWithdrawal(ctx context.Context, userID, currency string) (*models.WithdrawalRulesDTO, error) {
	rule := &models.WithdrawalRulesDTO{}
	data, err := o.storage.KeyValue().Get(ctx, formatKey(models.ExchangeSlugBybit, WithdrawalRuleType, currency, userID))
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, key_value.ErrEntryNotFound) {
			return o.fetchAndCacheBybitWithdrawalRules(ctx, userID, currency)
		}
		return nil, err
	}
	return rule, json.Unmarshal(data, rule)
}

/*

	Fetch and cache to key value storage

*/

func (o *Service) fetchAndCacheGateioWithdrawalRules(ctx context.Context, userID, currency string) (*models.WithdrawalRulesDTO, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	rules, err := o.fetchGateioWithdrawalRules(ctx, userUUID, currency)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return &models.WithdrawalRulesDTO{}, nil
	}
	err = o.storage.KeyValue().Set(ctx, formatKey(models.ExchangeSlugGateio, WithdrawalRuleType, currency, userID), structs.Map(rules[0]), 30*time.Minute)
	return rules[0], err
}

func (o *Service) fetchAndCacheKucoinWithdrawalRules(ctx context.Context, userID string, currency string) (*models.WithdrawalRulesDTO, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	rules, err := o.fetchKucoinWithdrawalRules(ctx, userUUID, currency)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return &models.WithdrawalRulesDTO{}, nil
	}
	err = o.storage.KeyValue().Set(ctx, formatKey(models.ExchangeSlugKucoin, WithdrawalRuleType, currency, userID), structs.Map(rules[0]), 30*time.Minute)
	return rules[0], err
}

func (o *Service) fetchAndCacheBitgetWithdrawalRules(ctx context.Context, userID string, currency string) (*models.WithdrawalRulesDTO, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	rules, err := o.fetchBitgetWithdrawalRules(ctx, userUUID, currency)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return &models.WithdrawalRulesDTO{}, nil
	}
	err = o.storage.KeyValue().Set(ctx, formatKey(models.ExchangeSlugBitget, WithdrawalRuleType, currency, userID), structs.Map(rules[0]), 30*time.Minute)
	return rules[0], err
}

func (o *Service) fetchAndCacheHtxWithdrawalRules(ctx context.Context, currency string) (*models.WithdrawalRulesDTO, error) {
	rules, err := o.fetchDefaultHtxWithdrawalRules(ctx, currency)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return &models.WithdrawalRulesDTO{}, nil
	}
	err = o.storage.KeyValue().Set(ctx, formatKey(models.ExchangeSlugHtx, WithdrawalRuleType, currency, "default"), structs.Map(rules[0]), 30*time.Minute)
	return rules[0], err
}

func (o *Service) fetchAndCacheOkxWithdrawalRules(ctx context.Context, userID, currency string) (*models.WithdrawalRulesDTO, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	rules, err := o.fetchOkxWithdrawalRules(ctx, userUUID, currency)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return &models.WithdrawalRulesDTO{}, nil
	}
	err = o.storage.KeyValue().Set(ctx, formatKey(models.ExchangeSlugOkx, WithdrawalRuleType, currency, userID), structs.Map(rules[0]), 30*time.Minute)
	return rules[0], err
}

func (o *Service) fetchAndCacheBinanceWithdrawalRules(ctx context.Context, userID, currency string) (*models.WithdrawalRulesDTO, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	rules, err := o.fetchBinanceWithdrawalRules(ctx, userUUID, currency)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return &models.WithdrawalRulesDTO{}, nil
	}
	err = o.storage.KeyValue().Set(ctx, formatKey(models.ExchangeSlugBinance, WithdrawalRuleType, currency, userID), structs.Map(rules[0]), 30*time.Minute)
	return rules[0], err
}

func (o *Service) fetchAndCacheBybitWithdrawalRules(ctx context.Context, userID, currency string) (*models.WithdrawalRulesDTO, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("parse user id: %w", err)
	}
	rules, err := o.fetchBybitWithdrawalRules(ctx, userUUID, currency)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return &models.WithdrawalRulesDTO{}, nil
	}
	err = o.storage.KeyValue().Set(ctx, formatKey(models.ExchangeSlugBybit, WithdrawalRuleType, currency, userID), structs.Map(rules[0]), 30*time.Minute)
	return rules[0], err
}

/*

	Update default exchange rules

*/

func (o *Service) updateHtxDefaultWithdrawalRules(ctx context.Context) error {
	enabledCurrencies, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugHtx)
	if err != nil {
		return err
	}
	publicExClient, err := o.manager.GetPublicDriver(ctx, models.ExchangeSlugHtx)
	if err != nil {
		return err
	}
	rules, err := publicExClient.GetWithdrawalRules(ctx, lo.Map(enabledCurrencies, func(c *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) string {
		return c.ID.String
	})...)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		id, exists := lo.Find(enabledCurrencies, func(c *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
			return c.Ticker == rule.Currency && c.Chain == rule.Chain
		})
		if !exists {
			return fmt.Errorf("currency %s not found in enabled currencies", rule.Currency)
		}
		err := o.storage.KeyValue().Set(ctx, formatKey(models.ExchangeSlugHtx, WithdrawalRuleType, id.ID.String, "default"), structs.Map(rule), 30*time.Minute)
		if err != nil {
			return fmt.Errorf("set exchange rule to storage: %w", err)
		}
	}
	return nil
}

/*

	Fetch exchange rules

*/

func (o *Service) fetchGateioWithdrawalRules(ctx context.Context, userID uuid.UUID, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	gateioClient, err := o.manager.GetDriver(ctx, models.ExchangeSlugGateio, userID)
	if err != nil {
		return nil, err
	}
	rules, err := gateioClient.GetWithdrawalRules(ctx, currencies...)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (o *Service) fetchKucoinWithdrawalRules(ctx context.Context, userID uuid.UUID, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	kucoinClient, err := o.manager.GetDriver(ctx, models.ExchangeSlugKucoin, userID)
	if err != nil {
		return nil, err
	}
	rules, err := kucoinClient.GetWithdrawalRules(ctx, currencies...)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (o *Service) fetchBitgetWithdrawalRules(ctx context.Context, userID uuid.UUID, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	bitgetClient, err := o.manager.GetDriver(ctx, models.ExchangeSlugBitget, userID)
	if err != nil {
		return nil, err
	}
	rules, err := bitgetClient.GetWithdrawalRules(ctx, currencies...)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (o *Service) fetchOkxWithdrawalRules(ctx context.Context, userID uuid.UUID, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	okxClient, err := o.manager.GetDriver(ctx, models.ExchangeSlugOkx, userID)
	if err != nil {
		return nil, err
	}
	rules, err := okxClient.GetWithdrawalRules(ctx, currencies...)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (o *Service) fetchBinanceWithdrawalRules(ctx context.Context, userID uuid.UUID, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	binanceClient, err := o.manager.GetDriver(ctx, models.ExchangeSlugBinance, userID)
	if err != nil {
		return nil, err
	}
	rules, err := binanceClient.GetWithdrawalRules(ctx, currencies...)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (o *Service) fetchBybitWithdrawalRules(ctx context.Context, userID uuid.UUID, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	bybitClient, err := o.manager.GetDriver(ctx, models.ExchangeSlugBybit, userID)
	if err != nil {
		return nil, err
	}
	rules, err := bybitClient.GetWithdrawalRules(ctx, currencies...)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

/*

	Fetch default exchange rules for exchanges with constant rules

*/

func (o *Service) fetchDefaultHtxWithdrawalRules(ctx context.Context, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	publicExClient, err := o.manager.GetPublicDriver(ctx, models.ExchangeSlugHtx)
	if err != nil {
		return nil, err
	}
	return publicExClient.GetWithdrawalRules(ctx, currencies...)
}

func (o *Service) updateUsersWithdrawalRules(ctx context.Context) error {
	users, err := o.storage.Users().GetAllWithExchangeEnabled(ctx)
	if err != nil {
		return err
	}
	for _, user := range users {
		if *user.ExchangeSlug == models.ExchangeSlugHtx {
			continue
		}
		err := o.updateUserWithdrawalRules(ctx, user)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Service) updateUserWithdrawalRules(ctx context.Context, user *models.User) error {
	exClient, err := o.manager.GetDriver(ctx, *user.ExchangeSlug, user.ID)
	if err != nil {
		return err
	}
	enabledCurrencies, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, *user.ExchangeSlug)
	if err != nil {
		return err
	}
	rules, err := exClient.GetWithdrawalRules(ctx, lo.Map(enabledCurrencies, func(c *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) string {
		return c.ID.String
	})...)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		id, exists := lo.Find(enabledCurrencies, func(c *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
			return c.Ticker == rule.Currency && c.Chain == rule.Chain
		})
		if !exists {
			return fmt.Errorf("currency %s not found in enabled currencies", rule.Currency)
		}
		err := o.storage.KeyValue().Set(ctx, formatKey(*user.ExchangeSlug, WithdrawalRuleType, id.ID.String, user.ID.String()), structs.Map(rule), 30*time.Minute)
		if err != nil {
			return fmt.Errorf("set exchange rule to storage: %w", err)
		}
	}
	return nil
}
