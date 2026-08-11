package mexc

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/mexc/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/mexc/responses"
)

const orderEndpoint = "/api/v3/order"

const (
	OrderSideBuy  = "BUY"
	OrderSideSell = "SELL"
)

type IMexcSpot interface {
	PlaceOrder(ctx context.Context, req *requests.PlaceOrderRequest) (*responses.PlaceOrderResponse, error)
	QueryOrder(ctx context.Context, req *requests.QueryOrderRequest) (*responses.QueryOrderResponse, error)
}

var _ IMexcSpot = (*SpotClient)(nil)

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
		orderEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 2, Period: time.Second}),
	}
}

func (o *SpotClient) PlaceOrder(ctx context.Context, req *requests.PlaceOrderRequest) (*responses.PlaceOrderResponse, error) {
	params := map[string]string{
		"symbol": req.Symbol,
		"side":   req.Side,
		"type":   req.Type,
	}
	if req.Quantity != "" {
		params["quantity"] = req.Quantity
	}
	if req.QuoteOrderQty != "" {
		params["quoteOrderQty"] = req.QuoteOrderQty
	}
	if req.NewClientOrderID != "" {
		params["newClientOrderId"] = req.NewClientOrderID
	}

	res := &responses.PlaceOrderResponse{}
	if err := o.client.Do(ctx, http.MethodPost, orderEndpoint, true, res, params); err != nil {
		return nil, err
	}
	return res, nil
}

func (o *SpotClient) QueryOrder(ctx context.Context, req *requests.QueryOrderRequest) (*responses.QueryOrderResponse, error) {
	params := map[string]string{"symbol": req.Symbol}
	if req.OrderID != "" {
		params["orderId"] = req.OrderID
	}
	if req.OrigClientOrderID != "" {
		params["origClientOrderId"] = req.OrigClientOrderID
	}

	res := &responses.QueryOrderResponse{}
	if err := o.client.Do(ctx, http.MethodGet, orderEndpoint, true, res, params); err != nil {
		return nil, err
	}
	return res, nil
}
