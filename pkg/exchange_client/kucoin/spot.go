package kucoin

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/ulule/limiter/v3"

	kucoinrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/requests"
	kucoinresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/responses"
)

const (
	createOrder         = "/api/v1/hf/orders"
	getOrderByOrderID   = "/api/v1/hf/orders/{orderId}"
	getOrderByClientOID = "/api/v1/hf/orders/client-order/{clientOid}"
)

var _ = IKucoinSpot((*Spot)(nil))

type IKucoinSpot interface {
	CreateOrder(ctx context.Context, req kucoinrequests.CreateOrder) (*kucoinresponses.CreateOrder, error)
	GetOrderByOrderID(ctx context.Context, req kucoinrequests.GetOrderByOrderID) (*kucoinresponses.GetOrderByOrderID, error)
	GetOrderByClientOID(ctx context.Context, req kucoinrequests.GetOrderByClientOID) (*kucoinresponses.GetOrderByClientOID, error)
}

type Spot struct {
	client *Client
}

func NewSpot(clientOpt *ClientOptions, store limiter.Store, opts ...ClientOption) *Spot {
	spot := &Spot{
		client: NewClient(clientOpt, store, opts...),
	}
	spot.initLimiters()
	return spot
}

func (o *Spot) initLimiters() {
	// KuCoin private endpoints: VIP0 allows 4000 requests per 30 seconds for Spot/Margin pool
	// Conservative approach: 80 requests per 30 seconds for trading operations
	// Order creation gets slightly lower limit for safety
	orderRate := limiter.Rate{Limit: 60, Period: 30 * time.Second}
	queryRate := limiter.Rate{Limit: 80, Period: 30 * time.Second}

	o.client.limiters = map[string]*limiter.Limiter{
		createOrder:         limiter.New(o.client.store, orderRate),
		getOrderByOrderID:   limiter.New(o.client.store, queryRate),
		getOrderByClientOID: limiter.New(o.client.store, queryRate),
	}
}

func (o *Spot) CreateOrder(ctx context.Context, req kucoinrequests.CreateOrder) (*kucoinresponses.CreateOrder, error) {
	response := &kucoinresponses.CreateOrder{}
	p := createOrder
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodPost, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Spot) GetOrderByOrderID(ctx context.Context, req kucoinrequests.GetOrderByOrderID) (*kucoinresponses.GetOrderByOrderID, error) {
	response := &kucoinresponses.GetOrderByOrderID{}
	p := getOrderByOrderID
	m := S2M(req)
	if req.OrderID != "" {
		p = strings.ReplaceAll(p, "{orderId}", req.OrderID)
	}
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Spot) GetOrderByClientOID(ctx context.Context, req kucoinrequests.GetOrderByClientOID) (*kucoinresponses.GetOrderByClientOID, error) {
	response := &kucoinresponses.GetOrderByClientOID{}
	p := getOrderByOrderID
	m := S2M(req)
	if req.ClientOrderOID != "" {
		p = strings.ReplaceAll(p, "{clientOid}", req.ClientOrderOID)
	}
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
