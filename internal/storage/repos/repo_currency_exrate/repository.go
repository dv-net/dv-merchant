package repo_currency_exrate

import (
	"context"
	"time"

	"github.com/dv-net/dv-merchant/pkg/key_value"
)

type ICurrencyRateRepo interface {
	StoreRate(ctx context.Context, from, to, source, value string, ttl time.Duration) error
	GetRate(ctx context.Context, from, to, source string) (string, error)
}

type Repo struct {
	driver key_value.IKeyValue
}

func (r *Repo) StoreRate(ctx context.Context, from, to, source, value string, ttl time.Duration) error {
	return r.driver.Set(ctx, source+":"+from+":"+to, value, ttl)
}

func (r *Repo) GetRate(ctx context.Context, from, to, source string) (string, error) {
	res, err := r.driver.Get(ctx, source+":"+from+":"+to)
	if err != nil {
		return "", err
	}
	return res.String(), nil
}

func New(driver key_value.IKeyValue) *Repo {
	return &Repo{
		driver: driver,
	}
}
