package htx

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	htxrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/requests"
	htxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/responses"
)

const (
	getOrderDetailsEndpoint           = "/v1/order/orders/%d"
	getOrderDetailsByClientIDEndpoint = "/v1/order/orders/getClientOrder"
	getOrderHistoryEndpoint           = "/v1/order/orders"
	postOrderPlaceEndpoint            = "/v1/order/orders/place"
)

type IHTXOrder interface {
	GetOrderDetails(ctx context.Context, orderID int64) (*htxresponses.GetOrder, error)
	GetOrderDetailsByClientID(ctx context.Context, dto *htxrequests.GetOrderByClientIDRequest) (*htxresponses.GetOrder, error)
	GetOrdersHistory(ctx context.Context, dto *htxrequests.GetOrderHistoryRequest) (*htxresponses.GetOrdersHistory, error)
	PlaceOrder(ctx context.Context, dto *htxrequests.PlaceOrderRequest) (*htxresponses.PlaceOrder, error)
}

type OrderClient struct {
	client *Client
}

func (o *OrderClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getOrderDetailsEndpoint:           limiter.New(o.client.store, limiter.Rate{Limit: 50, Period: 2 * time.Second}),
		getOrderDetailsByClientIDEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 50, Period: 2 * time.Second}),
		getOrderHistoryEndpoint:           limiter.New(o.client.store, limiter.Rate{Limit: 50, Period: 2 * time.Second}),
		postOrderPlaceEndpoint:            limiter.New(o.client.store, limiter.Rate{Limit: 100, Period: 2 * time.Second}),
	}
}

func (o *OrderClient) GetOrderDetails(ctx context.Context, orderID int64) (*htxresponses.GetOrder, error) {
	response := &htxresponses.GetOrder{}
	p := fmt.Sprintf(getOrderDetailsEndpoint, orderID)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *OrderClient) GetOrderDetailsByClientID(ctx context.Context, dto *htxrequests.GetOrderByClientIDRequest) (*htxresponses.GetOrder, error) {
	response := &htxresponses.GetOrder{}
	p := getOrderDetailsByClientIDEndpoint
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *OrderClient) GetOrdersHistory(ctx context.Context, dto *htxrequests.GetOrderHistoryRequest) (*htxresponses.GetOrdersHistory, error) {
	response := &htxresponses.GetOrdersHistory{}
	p := getOrderHistoryEndpoint
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *OrderClient) PlaceOrder(ctx context.Context, dto *htxrequests.PlaceOrderRequest) (*htxresponses.PlaceOrder, error) {
	response := &htxresponses.PlaceOrder{}
	p := postOrderPlaceEndpoint
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodPost, p, true, &response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func NewOrderClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *OrderClient {
	order := &OrderClient{
		client: NewClient(opt, store, opts...),
	}
	order.initLimiters()
	return order
}
