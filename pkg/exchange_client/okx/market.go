package okx

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	okxrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/requests"
	okxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/responses"
)

const (
	marketTickersEndpoint      = "/api/v5/market/tickers"
	marketTickerEndpoint       = "/api/v5/market/ticker"
	marketIndexTickersEndpoint = "/api/v5/market/index-tickers"
)

type IOKXMarket interface {
	GetTickers(ctx context.Context, req okxrequests.GetTickers) (*okxresponses.Ticker, error)
	GetTicker(ctx context.Context, req okxrequests.GetTicker) (*okxresponses.Ticker, error)
	GetIndexTickers(ctx context.Context, req okxrequests.GetIndexTickers) (*okxresponses.IndexTicker, error)
}

type Market struct {
	client   *Client
	store    limiter.Store
	limiters map[string]*limiter.Limiter
}

func NewMarket(clientOpt *ClientOptions, store limiter.Store, opts ...ClientOption) *Market {
	market := &Market{
		client: NewClient(clientOpt, store, opts...),
	}
	market.initLimiters()
	return market
}

func (o *Market) initLimiters() {
	o.limiters = map[string]*limiter.Limiter{
		marketTickersEndpoint:      limiter.New(o.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		marketTickerEndpoint:       limiter.New(o.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		marketIndexTickersEndpoint: limiter.New(o.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
	}
}

func (o *Market) GetIndexTickers(ctx context.Context, req okxrequests.GetIndexTickers) (*okxresponses.IndexTicker, error) {
	response := &okxresponses.IndexTicker{}
	p := marketIndexTickersEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, false, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Market) GetTickers(ctx context.Context, req okxrequests.GetTickers) (*okxresponses.Ticker, error) {
	response := &okxresponses.Ticker{}
	p := marketTickersEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, false, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Market) GetTicker(ctx context.Context, req okxrequests.GetTicker) (*okxresponses.Ticker, error) {
	response := &okxresponses.Ticker{}
	p := marketTickerEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, false, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
