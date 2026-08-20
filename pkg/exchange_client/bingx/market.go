package bingx

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bingx/responses"
)

const (
	getSymbolsEndpoint = "/openApi/spot/v1/common/symbols"
	getTickerEndpoint  = "/openApi/spot/v1/ticker/24hr"
)

type IBingxMarket interface {
	GetSymbols(ctx context.Context, symbol string) (*responses.GetSymbolsResponse, error)
	GetTicker(ctx context.Context, symbol string) (responses.GetTickerResponse, error)
}

var _ IBingxMarket = (*MarketClient)(nil)

type MarketClient struct {
	client *Client
}

func NewMarketClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *MarketClient {
	market := &MarketClient{
		client: NewClient(opt, store, opts...),
	}
	market.initLimiters()
	return market
}

func (o *MarketClient) initLimiters() {
	o.client.intervals = map[string]time.Duration{
		getSymbolsEndpoint: 200 * time.Millisecond,
		getTickerEndpoint:  200 * time.Millisecond,
	}
}

func (o *MarketClient) GetSymbols(ctx context.Context, symbol string) (*responses.GetSymbolsResponse, error) {
	params := map[string]string{}
	if symbol != "" {
		params["symbol"] = symbol
	}

	res := &responses.GetSymbolsResponse{}
	if err := o.client.Do(ctx, http.MethodGet, getSymbolsEndpoint, false, res, params); err != nil {
		return nil, err
	}
	return res, nil
}

func (o *MarketClient) GetTicker(ctx context.Context, symbol string) (responses.GetTickerResponse, error) {
	params := map[string]string{}
	if symbol != "" {
		params["symbol"] = symbol
	}

	res := responses.GetTickerResponse{}
	if err := o.client.Do(ctx, http.MethodGet, getTickerEndpoint, false, &res, params); err != nil {
		return nil, err
	}
	return res, nil
}
