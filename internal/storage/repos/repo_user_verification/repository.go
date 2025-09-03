package repo_user_verification

import (
	"context"
	"time"

	"github.com/dv-net/dv-merchant/pkg/key_value"
)

type IUserVerification interface {
	GetCodeByEmail(ctx context.Context, email string) (string, error)
	StoreCodeByEmail(ctx context.Context, email, code string, ttl time.Duration) error
	RemoveCode(ctx context.Context, email string) error
}

var _ IUserVerification = (*Repo)(nil)

func New(driver key_value.IKeyValue) *Repo {
	return &Repo{
		driver,
	}
}

type Repo struct {
	driver key_value.IKeyValue
}

func (r *Repo) GetCodeByEmail(ctx context.Context, email string) (string, error) {
	code, err := r.driver.Get(ctx, email)
	if err != nil {
		return "", err
	}
	return code.String(), nil
}

func (r *Repo) StoreCodeByEmail(ctx context.Context, email, code string, ttl time.Duration) error {
	return r.driver.Set(ctx, email, code, ttl)
}

func (r *Repo) RemoveCode(ctx context.Context, email string) error {
	return r.driver.Delete(ctx, email)
}
