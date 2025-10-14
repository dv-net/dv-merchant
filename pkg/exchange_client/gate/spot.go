package gateio

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"
)

const (
	getSpotCurrenciesEndpoint         = "/api/v4/spot/currencies"
	getSpotCurrencyEndpoint           = "/api/v4/spot/currencies/%s"
	getCurrencyPairsSupportedEndpoint = "/api/v4/spot/currency_pairs"
	getCurrencyPairSupportedEndpoint  = "/api/v4/spot/currency_pairs/%s"
	getSpotTickersEndpoint            = "/api/v4/spot/tickers"
	getSpotAccountBalancesEndpoint    = "/api/v4/spot/accounts"
	createSpotOrderEndpoint           = "/api/v4/spot/orders"
	getSpotOrderEndpoint              = "/api/v4/spot/orders/%s"
)

var _ IGateSpot = (*SpotClient)(nil)

type IGateSpot interface {
	GetSpotCurrencies(ctx context.Context) (*GetSpotCurrenciesResponse, error)
	GetSpotCurrency(ctx context.Context, currency string) (*GetSpotCurrencyResponse, error)
	GetSpotSupportedCurrencyPairs(ctx context.Context) (*GetSpotSupportedCurrencyPairsResponse, error)
	GetSpotSupportedCurrencyPair(ctx context.Context, currencyPair string) (*GetSpotSupportedCurrencyPairResponse, error)
	GetTickersInfo(ctx context.Context, dto *GetTickersInfoRequest) (*GetTickersInfoResponse, error)
	GetSpotAccountBalances(ctx context.Context, dto *GetSpotAccountBalancesRequest) (*GetSpotAccountBalancesResponse, error)
	CreateSpotOrder(ctx context.Context, dto *CreateSpotOrderRequest) (*CreateSpotOrderResponse, error)
	GetSpotOrder(ctx context.Context, orderID string) (*GetSpotOrderResponse, error)
}

type SpotClient struct {
	client *Client
}

func NewSpotClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *SpotClient {
	spot := &SpotClient{
		client: NewClient(opt, store, opts...),
	}
	spot.initLimiters()
	return spot
}

func (o *SpotClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getSpotCurrenciesEndpoint:         limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		getSpotCurrencyEndpoint:           limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		getCurrencyPairSupportedEndpoint:  limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		getCurrencyPairsSupportedEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
	}
}

func (o *SpotClient) GetSpotCurrencies(ctx context.Context) (*GetSpotCurrenciesResponse, error) {
	p := getSpotCurrenciesEndpoint
	response := &GetSpotCurrenciesResponse{}
	err := o.client.Do(ctx, http.MethodGet, p, false, &response.Data, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *SpotClient) GetSpotCurrency(ctx context.Context, currency string) (*GetSpotCurrencyResponse, error) {
	p := fmt.Sprintf(getSpotCurrencyEndpoint, currency)
	response := &GetSpotCurrencyResponse{}
	err := o.client.Do(ctx, http.MethodGet, p, false, &response.Data, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *SpotClient) GetSpotSupportedCurrencyPairs(ctx context.Context) (*GetSpotSupportedCurrencyPairsResponse, error) {
	p := getCurrencyPairsSupportedEndpoint
	response := &GetSpotSupportedCurrencyPairsResponse{}
	err := o.client.Do(ctx, http.MethodGet, p, false, &response.Data, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *SpotClient) GetSpotSupportedCurrencyPair(ctx context.Context, currencyPair string) (*GetSpotSupportedCurrencyPairResponse, error) {
	p := fmt.Sprintf(getCurrencyPairSupportedEndpoint, currencyPair)
	response := &GetSpotSupportedCurrencyPairResponse{}
	err := o.client.Do(ctx, http.MethodGet, p, false, &response.Data, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *SpotClient) GetTickersInfo(ctx context.Context, dto *GetTickersInfoRequest) (*GetTickersInfoResponse, error) {
	p := getSpotTickersEndpoint
	response := &GetTickersInfoResponse{}
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, false, &response.Data, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *SpotClient) GetSpotAccountBalances(ctx context.Context, dto *GetSpotAccountBalancesRequest) (*GetSpotAccountBalancesResponse, error) {
	p := getSpotAccountBalancesEndpoint
	response := &GetSpotAccountBalancesResponse{}
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response.Data, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *SpotClient) CreateSpotOrder(ctx context.Context, dto *CreateSpotOrderRequest) (*CreateSpotOrderResponse, error) {
	p := createSpotOrderEndpoint
	response := &CreateSpotOrderResponse{}
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodPost, p, true, &response.Data, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *SpotClient) GetSpotOrder(ctx context.Context, orderID string) (*GetSpotOrderResponse, error) {
	p := fmt.Sprintf(getSpotOrderEndpoint, orderID)
	response := &GetSpotOrderResponse{}
	err := o.client.Do(ctx, http.MethodGet, p, true, &response.Data, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}
