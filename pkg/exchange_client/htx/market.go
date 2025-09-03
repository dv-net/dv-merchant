package htx

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	htxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/responses"
)

const (
	getMarketTickersEndpoint = "/market/tickers"
	getMarketDetailEndpoint  = "/market/detail"
)

type IHTXMarket interface {
	GetMarketTickers(ctx context.Context) (*htxresponses.GetMarketTickers, error)
	GetMarketDetails(ctx context.Context, symbol string) (*htxresponses.GetMarketDetails, error)
}

func NewMarketClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *MarketClient {
	market := &MarketClient{
		client: NewClient(opt, store, opts...),
	}
	market.initLimiters()
	return market
}

type MarketClient struct {
	client *Client
}

func (o *MarketClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getMarketTickersEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 30, Period: 3 * time.Second}),
		getMarketDetailEndpoint:  limiter.New(o.client.store, limiter.Rate{Limit: 30, Period: 3 * time.Second}),
	}
}

func (o *MarketClient) GetMarketTickers(ctx context.Context) (*htxresponses.GetMarketTickers, error) {
	response := &htxresponses.GetMarketTickers{}
	p := getMarketTickersEndpoint
	err := o.client.Do(ctx, http.MethodGet, p, false, &response, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *MarketClient) GetMarketDetails(ctx context.Context, symbol string) (*htxresponses.GetMarketDetails, error) {
	response := &htxresponses.GetMarketDetails{}
	p := getMarketDetailEndpoint
	m := S2M(map[string]string{"symbol": symbol})
	err := o.client.Do(ctx, http.MethodGet, p, false, &response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
