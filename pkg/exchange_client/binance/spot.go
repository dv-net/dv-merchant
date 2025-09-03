package binance

import (
	"context"
	"net/http"
	"net/url"

	binancerequests "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/requests"
	binanceresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/responses"
)

type ISpotClient interface {
	TestNewOrder(ctx context.Context, request *binancerequests.TestNewOrderRequest) (*binanceresponses.TestNewOrderResponse, error)
	NewOrder(ctx context.Context, request *binancerequests.NewOrderRequest) (*binanceresponses.NewOrderResponse, error)
	QueryOrder(ctx context.Context, request *binancerequests.QueryOrderRequest) (*binanceresponses.QueryOrderResponse, error)
	CancelOrder(ctx context.Context, request *binancerequests.CancelOrderRequest) (*binanceresponses.CancelOrderResponse, error)
	AccountInformation(ctx context.Context, request *binancerequests.AccountInformationRequest) (*binanceresponses.AccountInformationResponse, error)
}

func NewSpotClient(opt *ClientOptions) (ISpotClient, error) {
	client, err := NewClient(opt)
	if err != nil {
		return nil, err
	}
	spot := &SpotClient{
		client: client,
	}

	return spot, nil
}

type SpotClient struct {
	client *Client
}

func (o *SpotClient) TestNewOrder(ctx context.Context, request *binancerequests.TestNewOrderRequest) (*binanceresponses.TestNewOrderResponse, error) {
	response := &binanceresponses.TestNewOrderResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/api/v3/order/test")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelSigned, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *SpotClient) NewOrder(ctx context.Context, request *binancerequests.NewOrderRequest) (*binanceresponses.NewOrderResponse, error) {
	response := &binanceresponses.NewOrderResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/api/v3/order")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelSigned, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *SpotClient) QueryOrder(ctx context.Context, request *binancerequests.QueryOrderRequest) (*binanceresponses.QueryOrderResponse, error) {
	response := &binanceresponses.QueryOrderResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/api/v3/order")
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
	if err := o.client.Do(ctx, req, SecurityLevelSigned, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *SpotClient) CancelOrder(ctx context.Context, request *binancerequests.CancelOrderRequest) (*binanceresponses.CancelOrderResponse, error) {
	response := &binanceresponses.CancelOrderResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/api/v3/order")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelSigned, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *SpotClient) AccountInformation(ctx context.Context, request *binancerequests.AccountInformationRequest) (*binanceresponses.AccountInformationResponse, error) {
	response := &binanceresponses.AccountInformationResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/api/v3/account")
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
	if err := o.client.Do(ctx, req, SecurityLevelSigned, response); err != nil {
		return nil, err
	}
	return response, nil
}
