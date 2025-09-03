package gateio

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"
)

var _ IGateAccount = (*AccountClient)(nil)

type IGateAccount interface {
	GetAccountDetail(context.Context) (*GetAccountDetailResponse, error)
}

const (
	getAccountDetailEndpoint = "/api/v4/account/detail"
)

type AccountClient struct {
	client *Client
}

func NewAccountClient(opt *ClientOptions, store limiter.Store) *AccountClient {
	account := &AccountClient{
		client: NewClient(opt, store),
	}
	account.initLimiters()
	return account
}

func (o *AccountClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getAccountDetailEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
	}
}

func (o *AccountClient) GetAccountDetail(ctx context.Context) (*GetAccountDetailResponse, error) {
	p := getAccountDetailEndpoint
	response := &GetAccountDetailResponse{}
	err := o.client.Do(ctx, http.MethodGet, p, true, &response.Data, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}
