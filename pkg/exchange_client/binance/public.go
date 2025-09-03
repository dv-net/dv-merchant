package binance

import (
	"context"
	"net/http"
	"net/url"

	binancerequests "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/requests"
	binanceresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/responses"
)

type IMarketClient interface {
	GetPing(ctx context.Context) (bool, error)
	GetServerTime(ctx context.Context) (*binanceresponses.GetServerTimeResponse, error)
	GetExchangeInfo(ctx context.Context, request *binancerequests.GetExchangeInfoRequest) (*binanceresponses.GetExchangeInfoResponse, error)
	GetSymbolPriceTicker(ctx context.Context, request *binancerequests.GetSymbolPriceTickerRequest) (*binanceresponses.GetSymbolPriceTickerResponse, error)
	GetSymbolsPriceTicker(ctx context.Context, request *binancerequests.GetSymbolsPriceTickerRequest) (*binanceresponses.GetSymbolsPriceTickerResponse, error)
}

func NewMarketData(opt *ClientOptions) (IMarketClient, error) {
	// We use specific endpoint for reference data.
	mdp := "https://data-api.binance.vision"
	bURL, err := url.Parse(mdp)
	if err != nil {
		return nil, err
	}
	newOpt := &ClientOptions{
		APIKey:       opt.APIKey,
		SecretKey:    opt.SecretKey,
		BaseURL:      bURL,
		PublicClient: opt.PublicClient,
	}

	client, err := NewClient(newOpt)
	if err != nil {
		return nil, err
	}
	marketData := &MarketDataClient{
		client: client,
	}

	return marketData, nil
}

type MarketDataClient struct {
	client *Client
}

func (o *MarketDataClient) GetPing(ctx context.Context) (bool, error) {
	path, err := url.JoinPath(o.client.baseURL.String(), "/api/v3/ping")
	if err != nil {
		return false, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return false, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelNone, &binanceresponses.GetPingResponse{}); err != nil {
		return false, err
	}
	return true, nil
}

func (o *MarketDataClient) GetServerTime(ctx context.Context) (*binanceresponses.GetServerTimeResponse, error) {
	response := &binanceresponses.GetServerTimeResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/api/v3/time")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelNone, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *MarketDataClient) GetExchangeInfo(ctx context.Context, request *binancerequests.GetExchangeInfoRequest) (*binanceresponses.GetExchangeInfoResponse, error) {
	response := &binanceresponses.GetExchangeInfoResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/api/v3/exchangeInfo")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelNone, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *MarketDataClient) GetSymbolPriceTicker(ctx context.Context, request *binancerequests.GetSymbolPriceTickerRequest) (*binanceresponses.GetSymbolPriceTickerResponse, error) {
	response := &binanceresponses.GetSymbolPriceTickerResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/api/v3/ticker/price")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelNone, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *MarketDataClient) GetSymbolsPriceTicker(ctx context.Context, request *binancerequests.GetSymbolsPriceTickerRequest) (*binanceresponses.GetSymbolsPriceTickerResponse, error) {
	response := &binanceresponses.GetSymbolsPriceTickerResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/api/v3/ticker/price")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelNone, &response.Data); err != nil {
		return nil, err
	}
	return response, nil
}
