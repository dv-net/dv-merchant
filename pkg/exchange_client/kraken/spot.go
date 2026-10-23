package kraken

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/responses"
)

const (
	addOrderEndpoint    = "/0/private/AddOrder"
	queryOrdersEndpoint = "/0/private/QueryOrders"

	maxQueryOrdersTxIDs = 50
)

const (
	OrderSideBuy  = "buy"
	OrderSideSell = "sell"

	OrderTypeMarket = "market"
	OrderTypeLimit  = "limit"
)

var ErrInvalidOrderRequest = errors.New("kraken: invalid order request")

type IKrakenSpot interface {
	AddOrder(ctx context.Context, req *requests.AddOrderRequest) (*responses.AddOrderResponse, error)
	QueryOrders(ctx context.Context, req *requests.QueryOrdersRequest) (responses.QueryOrdersResponse, error)
}

var _ IKrakenSpot = (*SpotClient)(nil)

type SpotClient struct {
	client *Client
}

func NewSpotClient(client *Client) *SpotClient {
	return &SpotClient{client: client}
}

func (o *SpotClient) AddOrder(ctx context.Context, req *requests.AddOrderRequest) (*responses.AddOrderResponse, error) {
	params, err := addOrderParams(req)
	if err != nil {
		return nil, err
	}

	res, err := DoPrivate[responses.AddOrderResponse](ctx, o.client, addOrderEndpoint, params)
	if err != nil {
		return nil, err
	}
	if !req.Validate && len(res.TxID) == 0 {
		return nil, fmt.Errorf("kraken: order accepted without txid: %s", res.Descr.Order)
	}
	return &res, nil
}

func (o *SpotClient) QueryOrders(ctx context.Context, req *requests.QueryOrdersRequest) (responses.QueryOrdersResponse, error) {
	if req == nil || len(req.TxIDs) == 0 {
		return nil, fmt.Errorf("%w: at least one txid is required", ErrInvalidOrderRequest)
	}
	if len(req.TxIDs) > maxQueryOrdersTxIDs {
		return nil, fmt.Errorf("%w: at most %d txids per request", ErrInvalidOrderRequest, maxQueryOrdersTxIDs)
	}

	params := url.Values{"txid": {strings.Join(req.TxIDs, ",")}}
	if req.Trades {
		params.Set("trades", "true")
	}
	if req.UserRef != nil {
		params.Set("userref", strconv.FormatInt(int64(*req.UserRef), 10))
	}
	if req.ConsolidateTaker != nil {
		params.Set("consolidate_taker", strconv.FormatBool(*req.ConsolidateTaker))
	}

	return DoPrivate[responses.QueryOrdersResponse](ctx, o.client, queryOrdersEndpoint, params)
}

func addOrderParams(req *requests.AddOrderRequest) (url.Values, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is nil", ErrInvalidOrderRequest)
	}
	if req.Pair == "" {
		return nil, fmt.Errorf("%w: pair is required", ErrInvalidOrderRequest)
	}
	if req.Type != OrderSideBuy && req.Type != OrderSideSell {
		return nil, fmt.Errorf("%w: type must be %q or %q, got %q", ErrInvalidOrderRequest, OrderSideBuy, OrderSideSell, req.Type)
	}
	if req.OrderType == "" {
		return nil, fmt.Errorf("%w: ordertype is required", ErrInvalidOrderRequest)
	}
	if !req.Volume.IsPositive() {
		return nil, fmt.Errorf("%w: volume must be positive, got %s", ErrInvalidOrderRequest, req.Volume)
	}
	if req.OrderType == OrderTypeLimit && !req.Price.IsPositive() {
		return nil, fmt.Errorf("%w: price is required for limit order", ErrInvalidOrderRequest)
	}

	params := url.Values{
		"pair":      {req.Pair},
		"type":      {req.Type},
		"ordertype": {req.OrderType},
		"volume":    {req.Volume.String()},
	}
	if !req.Price.IsZero() {
		params.Set("price", req.Price.String())
	}
	if req.ClOrdID != "" {
		params.Set("cl_ord_id", req.ClOrdID)
	}
	if req.UserRef != nil {
		params.Set("userref", strconv.FormatInt(int64(*req.UserRef), 10))
	}
	if len(req.OFlags) > 0 {
		params.Set("oflags", strings.Join(req.OFlags, ","))
	}
	if req.TimeInForce != "" {
		params.Set("timeinforce", req.TimeInForce)
	}
	if req.Validate {
		params.Set("validate", "true")
	}
	return params, nil
}
