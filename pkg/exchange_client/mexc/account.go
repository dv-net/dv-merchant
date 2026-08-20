package mexc

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/mexc/responses"
)

const getAccountInfoEndpoint = "/api/v3/account"

type IMexcAccount interface {
	GetAccountInfo(ctx context.Context) (*responses.AccountInfoResponse, error)
}

var _ IMexcAccount = (*AccountClient)(nil)

type AccountClient struct {
	client *Client
}

func NewAccountClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *AccountClient {
	account := &AccountClient{
		client: NewClient(opt, store, opts...),
	}
	account.initLimiters()
	return account
}

func (o *AccountClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getAccountInfoEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 2, Period: time.Second}),
	}
}

func (o *AccountClient) GetAccountInfo(ctx context.Context) (*responses.AccountInfoResponse, error) {
	res := &responses.AccountInfoResponse{}
	if err := o.client.Do(ctx, http.MethodGet, getAccountInfoEndpoint, true, res); err != nil {
		return nil, err
	}
	return res, nil
}
