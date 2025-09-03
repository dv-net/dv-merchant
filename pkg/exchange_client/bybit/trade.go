package bybit

import (
	"context"
	"net/http"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/responses"

	"github.com/ulule/limiter/v3"
)

type ITradeClient interface {
	GetActiveOrders(ctx context.Context, req *requests.GetActiveOrdersRequest) (*responses.BaseResponse[responses.GetActiveOrdersResponse], error)
	GetOrderHistory(ctx context.Context, req *requests.GetOrderHistoryRequest) (*responses.BaseResponse[responses.GetOrderHistoryResponse], error)
	PlaceOrder(ctx context.Context, req *requests.PlaceOrderRequest) (*responses.BaseResponse[responses.PlaceOrderResponse], error)
}

type TradeClient struct {
	client *Client
}

func NewTrade(opt *ClientOptions, store limiter.Store, opts ...ClientOption) ITradeClient {
	client := NewClient(opt, store, opts...)
	return &TradeClient{
		client: client,
	}
}

func (o *TradeClient) GetActiveOrders(ctx context.Context, req *requests.GetActiveOrdersRequest) (*responses.BaseResponse[responses.GetActiveOrdersResponse], error) {
	endpoint := "/v5/order/realtime"

	params := S2M(req)
	var resp responses.BaseResponse[responses.GetActiveOrdersResponse]
	err := o.client.Do(ctx, http.MethodGet, endpoint, true, &resp, params)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *TradeClient) GetOrderHistory(ctx context.Context, req *requests.GetOrderHistoryRequest) (*responses.BaseResponse[responses.GetOrderHistoryResponse], error) {
	endpoint := "/v5/order/history"

	params := S2M(req)
	var resp responses.BaseResponse[responses.GetOrderHistoryResponse]
	err := o.client.Do(ctx, http.MethodGet, endpoint, true, &resp, params)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *TradeClient) PlaceOrder(ctx context.Context, req *requests.PlaceOrderRequest) (*responses.BaseResponse[responses.PlaceOrderResponse], error) {
	endpoint := "/v5/order/create"

	params := S2M(req)
	var resp responses.BaseResponse[responses.PlaceOrderResponse]
	err := o.client.Do(ctx, http.MethodPost, endpoint, true, &resp, params)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
