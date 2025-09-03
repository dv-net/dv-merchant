package clients

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/responses"
)

const (
	getTickerInformationEndpoint = "/api/v2/spot/market/tickers"
	getCoinInformationEndpoint   = "/api/v2/spot/public/coins"
	getSymbolInformationEndpoint = "/api/v2/spot/public/symbols"
)

var _ IBitgetMarket = (*MarketClient)(nil)

type IBitgetMarket interface {
	TickerInformation(context.Context, *requests.TickerInformationRequest) (*responses.TickerInformationResponse, error)
	CoinInformation(context.Context, *requests.CoinInformationRequest) (*responses.CoinInformationResponse, error)
	SymbolInformation(context.Context, *requests.SymbolInformationRequest) (*responses.SymbolInformationResponse, error)
}

func NewMarketClient(opt *ClientOptions, store limiter.Store, signer bitget.ISigner) *MarketClient {
	market := &MarketClient{
		client: NewClient(opt, store, signer),
	}
	market.initLimiters()
	return market
}

func (o *MarketClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getTickerInformationEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: time.Second}),
		getCoinInformationEndpoint:   limiter.New(o.client.store, limiter.Rate{Limit: 3, Period: time.Second}),
		getSymbolInformationEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: time.Second}),
	}
}

type MarketClient struct {
	client *Client
}

func (o *MarketClient) TickerInformation(ctx context.Context, req *requests.TickerInformationRequest) (*responses.TickerInformationResponse, error) {
	response := &responses.TickerInformationResponse{}
	p := getTickerInformationEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, false, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *MarketClient) CoinInformation(ctx context.Context, req *requests.CoinInformationRequest) (*responses.CoinInformationResponse, error) {
	response := &responses.CoinInformationResponse{}
	p := getCoinInformationEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, false, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *MarketClient) SymbolInformation(ctx context.Context, req *requests.SymbolInformationRequest) (*responses.SymbolInformationResponse, error) {
	response := &responses.SymbolInformationResponse{}
	p := getSymbolInformationEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, false, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
