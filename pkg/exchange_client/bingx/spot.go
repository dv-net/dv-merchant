package bingx

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bingx/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bingx/responses"
)

const (
	placeOrderEndpoint = "/openApi/spot/v1/trade/order"
	queryOrderEndpoint = "/openApi/spot/v1/trade/query"
)

const (
	OrderSideBuy  = "BUY"
	OrderSideSell = "SELL"

	OrderTypeMarket = "MARKET"
)

type IBingxSpot interface {
	PlaceOrder(ctx context.Context, req *requests.PlaceOrderRequest) (*responses.PlaceOrderResponse, error)
	QueryOrder(ctx context.Context, req *requests.QueryOrderRequest) (*responses.QueryOrderResponse, error)
}

var _ IBingxSpot = (*SpotClient)(nil)

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
	o.client.intervals = map[string]time.Duration{
		placeOrderEndpoint: 200 * time.Millisecond,
		queryOrderEndpoint: 200 * time.Millisecond,
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
	if req.ClientOrderID != "" {
		params["newClientOrderId"] = req.ClientOrderID
	}

	res := &responses.PlaceOrderResponse{}
	if err := o.client.Do(ctx, http.MethodPost, placeOrderEndpoint, true, res, params); err != nil {
		return nil, err
	}
	return res, nil
}

func (o *SpotClient) QueryOrder(ctx context.Context, req *requests.QueryOrderRequest) (*responses.QueryOrderResponse, error) {
	params := map[string]string{"symbol": req.Symbol}
	if req.OrderID != "" {
		params["orderId"] = req.OrderID
	}
	if req.ClientOrderID != "" {
		params["clientOrderID"] = req.ClientOrderID
	}

	res := &responses.QueryOrderResponse{}
	if err := o.client.Do(ctx, http.MethodGet, queryOrderEndpoint, true, res, params); err != nil {
		return nil, err
	}
	return res, nil
}
