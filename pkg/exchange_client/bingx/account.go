package bingx

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bingx/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bingx/responses"
)

const (
	getBalanceEndpoint     = "/openApi/spot/v1/account/balance"
	getFundBalanceEndpoint = "/openApi/fund/v1/account/balance"
	transferEndpoint       = "/openApi/api/v3/post/asset/transfer"
)

const (
	TransferFundToSpot = "FUND_SPOT"
	TransferSpotToFund = "SPOT_FUND"
)

type IBingxAccount interface {
	GetBalance(ctx context.Context) (*responses.GetBalanceResponse, error)
	GetFundBalance(ctx context.Context) (*responses.GetFundBalanceResponse, error)
	Transfer(ctx context.Context, req *requests.TransferRequest) (*responses.TransferResponse, error)
}

var _ IBingxAccount = (*AccountClient)(nil)

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
	o.client.intervals = map[string]time.Duration{
		getBalanceEndpoint:     400 * time.Millisecond,
		getFundBalanceEndpoint: 600 * time.Millisecond,
		transferEndpoint:       600 * time.Millisecond,
	}
}

func (o *AccountClient) GetBalance(ctx context.Context) (*responses.GetBalanceResponse, error) {
	res := &responses.GetBalanceResponse{}
	if err := o.client.Do(ctx, http.MethodGet, getBalanceEndpoint, true, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (o *AccountClient) GetFundBalance(ctx context.Context) (*responses.GetFundBalanceResponse, error) {
	res := &responses.GetFundBalanceResponse{}
	if err := o.client.Do(ctx, http.MethodGet, getFundBalanceEndpoint, true, res); err != nil {
		return nil, err
	}
	return res, nil
}

func (o *AccountClient) Transfer(ctx context.Context, req *requests.TransferRequest) (*responses.TransferResponse, error) {
	params := map[string]string{
		"type":   req.Type,
		"asset":  req.Asset,
		"amount": req.Amount,
	}

	res := &responses.TransferResponse{}
	if err := o.client.Do(ctx, http.MethodPost, transferEndpoint, true, res, params); err != nil {
		return nil, err
	}
	return res, nil
}
