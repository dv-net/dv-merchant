package mexc

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/mexc/responses"
)

const (
	getExchangeInfoEndpoint = "/api/v3/exchangeInfo"
	getTickerPriceEndpoint  = "/api/v3/ticker/price"
)

const (
	SymbolStatusEnabled = "1"
	SymbolStatusPaused  = "2"
	SymbolStatusOffline = "3"

	OrderTypeMarket = "MARKET"
)

const (
	TradeSideTypeAll      = 1
	TradeSideTypeBuyOnly  = 2
	TradeSideTypeSellOnly = 3
	TradeSideTypeClose    = 4
)

type IMexcMarket interface {
	GetExchangeInfo(ctx context.Context, symbol string) (*responses.GetExchangeInfoResponse, error)
	GetTickerPrice(ctx context.Context, symbol string) (*responses.GetTickerPriceResponse, error)
}

var _ IMexcMarket = (*MarketClient)(nil)

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
	o.client.limiters = map[string]*limiter.Limiter{
		getExchangeInfoEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 2, Period: time.Second}),
		getTickerPriceEndpoint:  limiter.New(o.client.store, limiter.Rate{Limit: 2, Period: time.Second}),
	}
}

func (o *MarketClient) GetExchangeInfo(ctx context.Context, symbol string) (*responses.GetExchangeInfoResponse, error) {
	params := map[string]string{}
	if symbol != "" {
		params["symbol"] = symbol
	}

	res := &responses.GetExchangeInfoResponse{}
	if err := o.client.Do(ctx, http.MethodGet, getExchangeInfoEndpoint, false, res, params); err != nil {
		return nil, err
	}
	return res, nil
}

func (o *MarketClient) GetTickerPrice(ctx context.Context, symbol string) (*responses.GetTickerPriceResponse, error) {
	res := &responses.GetTickerPriceResponse{}
	if err := o.client.Do(ctx, http.MethodGet, getTickerPriceEndpoint, false, res, map[string]string{"symbol": symbol}); err != nil {
		return nil, err
	}
	return res, nil
}
