package repo_currency_exrate

import (
	"context"
	"errors"
	"time"

	"github.com/dv-net/dv-merchant/pkg/key_value"
	"github.com/shopspring/decimal"
)

type ICurrencyRateRepo interface {
	StoreRate(ctx context.Context, from, to, source, value string, ttl time.Duration) error
	GetRate(ctx context.Context, from, to, source string) (string, error)
}

type Repo struct {
	driver key_value.IKeyValue
}

func (r *Repo) StoreRate(ctx context.Context, source, from, to, value string, ttl time.Duration) error {
	return r.driver.Set(ctx, source+":"+from+":"+to, value, ttl)
}

func (r *Repo) GetRate(ctx context.Context, source, from, to string) (string, error) {
	res, err := r.driver.Get(ctx, source+":"+from+":"+to)
	if err == nil {
		return res.String(), nil
	}

	if !errors.Is(err, key_value.ErrEntryNotFound) {
		return "", err
	}

	pattern := "*:" + from + ":" + to
	return r.getBestRate(ctx, pattern)
}

func (r *Repo) getBestRate(ctx context.Context, pattern string) (string, error) {
	keys, err := r.driver.Keys(ctx, pattern)

	if err != nil {
		return "", err
	}

	if len(keys) == 0 {
		return "", key_value.ErrEntryNotFound
	}

	var bestRate decimal.Decimal
	var bestRateStr string
	firstIteration := true

	for _, key := range keys {
		res, err := r.driver.Get(ctx, key)
		if err != nil {
			continue
		}

		rateStr := res.String()
		rate, err := decimal.NewFromString(rateStr)
		if err != nil {
			continue
		}

		if firstIteration || rate.LessThan(bestRate) {
			bestRate = rate
			bestRateStr = rateStr
			firstIteration = false
		}
	}

	if bestRateStr == "" {
		return "", key_value.ErrEntryNotFound
	}

	return bestRateStr, nil
}

func New(driver key_value.IKeyValue) *Repo {
	return &Repo{
		driver: driver,
	}
}
