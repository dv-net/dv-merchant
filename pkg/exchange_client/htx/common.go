package htx

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	htxrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/requests"
	htxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/responses"
)

const (
	getAllSupportedTradingSymbolsEndpoint = "/v2/settings/common/symbols"
	getAllSupportedCurrenciesEndpoint     = "/v2/settings/common/currencies"
	getAllMarketSymbols                   = "/v1/settings/common/market-symbols"
	getAllCommonSymbols                   = "/v1/common/symbols"
	getCurrencyReference                  = "/v2/reference/currencies"
)

type IHTXCommon interface {
	GetAllSupportedSymbols(ctx context.Context) (*htxresponses.GetSymbols, error)
	GetAllSupportedCurrencies(ctx context.Context, dto *htxrequests.GetAllSupportedCurrenciesRequest) (*htxresponses.GetCurrencies, error)
	GetAllCommonSymbols(ctx context.Context) (*htxresponses.GetCommonSymbols, error)
	GetAllMarketSymbols(ctx context.Context, dto *htxrequests.GetMarketSymbolsRequest) (*htxresponses.GetMarketSymbols, error)
	GetCurrencyReference(ctx context.Context, dto *htxrequests.GetCurrencyReferenceRequest) (*htxresponses.GetCurrencyReference, error)
}

func NewCommonClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *CommonClient {
	common := &CommonClient{
		client: NewClient(opt, store, opts...),
	}
	common.initLimiters()
	return common
}

type CommonClient struct {
	client *Client
}

func (o *CommonClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getAllSupportedTradingSymbolsEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 30, Period: 3 * time.Second}),
		getAllSupportedCurrenciesEndpoint:     limiter.New(o.client.store, limiter.Rate{Limit: 30, Period: 3 * time.Second}),
		getAllMarketSymbols:                   limiter.New(o.client.store, limiter.Rate{Limit: 30, Period: 3 * time.Second}),
		getAllCommonSymbols:                   limiter.New(o.client.store, limiter.Rate{Limit: 30, Period: 3 * time.Second}),
		getCurrencyReference:                  limiter.New(o.client.store, limiter.Rate{Limit: 60, Period: 3 * time.Second}),
	}
}

func (o *CommonClient) GetAllMarketSymbols(ctx context.Context, dto *htxrequests.GetMarketSymbolsRequest) (*htxresponses.GetMarketSymbols, error) {
	response := &htxresponses.GetMarketSymbols{}
	p := getAllMarketSymbols
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetAllCommonSymbols - deprecated
func (o *CommonClient) GetAllCommonSymbols(ctx context.Context) (*htxresponses.GetCommonSymbols, error) {
	response := &htxresponses.GetCommonSymbols{}
	p := getAllCommonSymbols
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *CommonClient) GetAllSupportedSymbols(ctx context.Context) (*htxresponses.GetSymbols, error) {
	response := &htxresponses.GetSymbols{}
	p := getAllSupportedTradingSymbolsEndpoint
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *CommonClient) GetAllSupportedCurrencies(ctx context.Context, dto *htxrequests.GetAllSupportedCurrenciesRequest) (*htxresponses.GetCurrencies, error) {
	response := &htxresponses.GetCurrencies{}
	p := getAllSupportedCurrenciesEndpoint
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *CommonClient) GetCurrencyReference(ctx context.Context, dto *htxrequests.GetCurrencyReferenceRequest) (*htxresponses.GetCurrencyReference, error) {
	response := &htxresponses.GetCurrencyReference{}
	p := getCurrencyReference
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, false, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
