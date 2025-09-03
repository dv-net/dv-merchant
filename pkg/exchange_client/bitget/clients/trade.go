package clients

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget"
	bitgetrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/requests"
	bitgetresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/responses"
)

const (
	getOrderInformationEndpoint = "/api/v2/spot/trade/orderInfo"
	postPlaceOrderEndpoint      = "/api/v2/spot/trade/place-order"
)

var _ IBitgetTrade = (*TradeClient)(nil)

type IBitgetTrade interface {
	OrderInformation(context.Context, *bitgetrequests.OrderInformationRequest) (*bitgetresponses.OrderInformationResponse, error)
	PlaceOrder(context.Context, *bitgetrequests.PlaceOrderRequest) (*bitgetresponses.PlaceOrderResponse, error)
}

func NewTradeClient(opt *ClientOptions, store limiter.Store, signer bitget.ISigner) *TradeClient {
	trade := &TradeClient{
		client: NewClient(opt, store, signer),
	}
	trade.initLimiters()
	return trade
}

type TradeClient struct {
	client *Client
}

func (o *TradeClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getOrderInformationEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: time.Second}),
	}
}

func (o *TradeClient) OrderInformation(ctx context.Context, req *bitgetrequests.OrderInformationRequest) (*bitgetresponses.OrderInformationResponse, error) {
	response := &bitgetresponses.OrderInformationResponse{}
	p := getOrderInformationEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *TradeClient) PlaceOrder(ctx context.Context, req *bitgetrequests.PlaceOrderRequest) (*bitgetresponses.PlaceOrderResponse, error) {
	response := &bitgetresponses.PlaceOrderResponse{}
	p := postPlaceOrderEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodPost, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
