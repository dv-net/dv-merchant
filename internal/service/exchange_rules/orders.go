package exchange_rules

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/pkg/key_value"

	"github.com/fatih/structs"
	"github.com/redis/go-redis/v9"
)

func (o *Service) GetSpotOrderRule(ctx context.Context, exchange models.ExchangeSlug, ticker string) (*models.OrderRulesDTO, error) {
	rules, err := o.getSpotOrderRules(ctx, exchange, ticker)
	if err != nil || len(rules) == 0 {
		return nil, fmt.Errorf("no rules found for ticker %s", ticker)
	}
	return rules[0], nil
}

/*

	Handlers

*/

func (o *Service) handleKucoinOrder(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	rule := &models.OrderRulesDTO{}
	data, err := o.storage.KeyValue().Get(ctx, formatKey(models.ExchangeSlugKucoin, SpotOrderRuleType, ticker, "default"))
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, key_value.ErrEntryNotFound) {
			return o.fetchAndCacheKucoinOrderRules(ctx, ticker)
		}
		return nil, err
	}
	if err := json.Unmarshal(data.Bytes(), rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (o *Service) handleBitgetOrder(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	rule := &models.OrderRulesDTO{}
	data, err := o.storage.KeyValue().Get(ctx, formatKey(models.ExchangeSlugBitget, SpotOrderRuleType, ticker, "default"))
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, key_value.ErrEntryNotFound) {
			return o.fetchAndCacheBitgetOrderRules(ctx, ticker)
		}
		return nil, err
	}
	if err := json.Unmarshal(data.Bytes(), rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (o *Service) handleHtxOrder(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	rule := &models.OrderRulesDTO{}
	data, err := o.storage.KeyValue().Get(ctx, formatKey(models.ExchangeSlugHtx, SpotOrderRuleType, ticker, "default"))
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, key_value.ErrEntryNotFound) {
			return o.fetchAndCacheHtxOrderRules(ctx, ticker)
		}
		return nil, err
	}
	if err := json.Unmarshal(data.Bytes(), rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (o *Service) handleOkxOrder(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	rule := &models.OrderRulesDTO{}
	data, err := o.storage.KeyValue().Get(ctx, formatKey(models.ExchangeSlugOkx, SpotOrderRuleType, ticker, "default"))
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, key_value.ErrEntryNotFound) {
			return o.fetchAndCacheOkxOrderRules(ctx, ticker)
		}
		return nil, err
	}
	if err := json.Unmarshal(data.Bytes(), rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (o *Service) handleBinanceOrder(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	rule := &models.OrderRulesDTO{}
	data, err := o.storage.KeyValue().Get(ctx, formatKey(models.ExchangeSlugBinance, SpotOrderRuleType, ticker, "default"))
	if err != nil {
		if errors.Is(err, redis.Nil) || errors.Is(err, key_value.ErrEntryNotFound) {
			return o.fetchAndCacheBinanceOrderRules(ctx, ticker)
		}
		return nil, err
	}
	if err := json.Unmarshal(data, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

/*

	End handlers

*/

func (o *Service) getSpotOrderRules(ctx context.Context, exchange models.ExchangeSlug, tickers ...string) ([]*models.OrderRulesDTO, error) {
	rules := make([]*models.OrderRulesDTO, 0, len(tickers))
	for _, currency := range tickers {
		rule, err := o.fetchSpotOrderRule(ctx, exchange, currency)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (o *Service) fetchSpotOrderRule(ctx context.Context, exchange models.ExchangeSlug, ticker string) (*models.OrderRulesDTO, error) {
	switch exchange {
	case models.ExchangeSlugBitget:
		return o.handleBitgetOrder(ctx, ticker)
	case models.ExchangeSlugHtx:
		return o.handleHtxOrder(ctx, ticker)
	case models.ExchangeSlugOkx:
		return o.handleOkxOrder(ctx, ticker)
	case models.ExchangeSlugBinance:
		return o.handleBinanceOrder(ctx, ticker)
	case models.ExchangeSlugKucoin:
		return o.handleKucoinOrder(ctx, ticker)
	default:
		return nil, fmt.Errorf("exchange %s not supported", exchange.String())
	}
}

func (o *Service) fetchAndCacheKucoinOrderRules(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	rules, err := o.fetchDefaultKucoinOrderRules(ctx, ticker)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return &models.OrderRulesDTO{}, nil
	}
	err = o.storage.KeyValue().Set(ctx, formatKey(models.ExchangeSlugKucoin, SpotOrderRuleType, ticker, "default"), structs.Map(rules[0]), 30*time.Minute)
	return rules[0], err
}

func (o *Service) fetchDefaultKucoinOrderRules(ctx context.Context, tickers ...string) ([]*models.OrderRulesDTO, error) {
	publicExClient, err := o.manager.GetPublicDriver(ctx, models.ExchangeSlugKucoin)
	if err != nil {
		return nil, err
	}
	return publicExClient.GetOrderRules(ctx, tickers...)
}

func (o *Service) fetchAndCacheBitgetOrderRules(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	rules, err := o.fetchDefaultBitgetOrderRules(ctx, ticker)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return &models.OrderRulesDTO{}, nil
	}
	err = o.storage.KeyValue().Set(ctx, formatKey(models.ExchangeSlugBitget, SpotOrderRuleType, ticker, "default"), structs.Map(rules[0]), 30*time.Minute)
	return rules[0], err
}

func (o *Service) fetchDefaultBitgetOrderRules(ctx context.Context, tickers ...string) ([]*models.OrderRulesDTO, error) {
	publicExClient, err := o.manager.GetPublicDriver(ctx, models.ExchangeSlugBitget)
	if err != nil {
		return nil, err
	}
	return publicExClient.GetOrderRules(ctx, tickers...)
}

func (o *Service) fetchAndCacheBinanceOrderRules(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	rules, err := o.fetchDefaultBinanceOrderRules(ctx, ticker)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return &models.OrderRulesDTO{}, nil
	}
	err = o.storage.KeyValue().Set(ctx, formatKey(models.ExchangeSlugBinance, SpotOrderRuleType, ticker, "default"), structs.Map(rules[0]), 30*time.Minute)
	return rules[0], err
}

func (o *Service) fetchDefaultBinanceOrderRules(ctx context.Context, tickers ...string) ([]*models.OrderRulesDTO, error) {
	publicExClient, err := o.manager.GetPublicDriver(ctx, models.ExchangeSlugBinance)
	if err != nil {
		return nil, err
	}
	return publicExClient.GetOrderRules(ctx, tickers...)
}

func (o *Service) fetchAndCacheOkxOrderRules(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	rules, err := o.fetchDefaultOkxOrderRules(ctx, ticker)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return &models.OrderRulesDTO{}, nil
	}
	err = o.storage.KeyValue().Set(ctx, formatKey(models.ExchangeSlugOkx, SpotOrderRuleType, ticker, "default"), structs.Map(rules[0]), 30*time.Minute)
	return rules[0], err
}

func (o *Service) fetchDefaultOkxOrderRules(ctx context.Context, tickers ...string) ([]*models.OrderRulesDTO, error) {
	publicExClient, err := o.manager.GetPublicDriver(ctx, models.ExchangeSlugOkx)
	if err != nil {
		return nil, err
	}
	return publicExClient.GetOrderRules(ctx, tickers...)
}

func (o *Service) fetchAndCacheHtxOrderRules(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	rules, err := o.fetchDefaultHtxOrderRules(ctx, ticker)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return &models.OrderRulesDTO{}, nil
	}
	err = o.storage.KeyValue().Set(ctx, formatKey(models.ExchangeSlugHtx, SpotOrderRuleType, ticker, "default"), structs.Map(rules[0]), 30*time.Minute)
	return rules[0], err
}

func (o *Service) fetchDefaultHtxOrderRules(ctx context.Context, tickers ...string) ([]*models.OrderRulesDTO, error) {
	publicExClient, err := o.manager.GetPublicDriver(ctx, models.ExchangeSlugHtx)
	if err != nil {
		return nil, err
	}
	return publicExClient.GetOrderRules(ctx, tickers...)
}
