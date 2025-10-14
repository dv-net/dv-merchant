package clients

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget"
	bitgetresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/responses"
)

const (
	getAllAccountBalanceEndpoint = "/api/v2/account/all-account-balance"
	getServerTimeEndpoint        = "/api/v2/public/time"
)

var _ IBitgetCommon = (*CommonClient)(nil)

type IBitgetCommon interface {
	AllAccountBalance(context.Context) (*bitgetresponses.AllAccountBalanceResponse, error)
	ServerTime(context.Context) (*bitgetresponses.ServerTimeResponse, error)
}

func NewCommonClient(opt *ClientOptions, store limiter.Store, signer bitget.ISigner, opts ...SubClientOption) *CommonClient {
	common := &CommonClient{
		client: NewClient(opt, store, signer, opts...),
	}
	common.initLimiters()
	return common
}

type CommonClient struct {
	client *Client
}

func (o *CommonClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getAllAccountBalanceEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 1, Period: time.Second}),
		getServerTimeEndpoint:        limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: time.Second}),
	}
}

func (o *CommonClient) AllAccountBalance(ctx context.Context) (*bitgetresponses.AllAccountBalanceResponse, error) {
	response := &bitgetresponses.AllAccountBalanceResponse{}
	p := getAllAccountBalanceEndpoint
	err := o.client.Do(ctx, http.MethodGet, p, true, response, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *CommonClient) ServerTime(ctx context.Context) (*bitgetresponses.ServerTimeResponse, error) {
	response := &bitgetresponses.ServerTimeResponse{}
	p := getServerTimeEndpoint
	err := o.client.Do(ctx, http.MethodGet, p, false, response, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}
