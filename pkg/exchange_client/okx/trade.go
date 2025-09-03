package okx

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	okxrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/requests"
	okxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/responses"
)

type IOKXTrade interface {
	PlaceOrder(ctx context.Context, req []okxrequests.PlaceOrder) (*okxresponses.PlaceOrder, error)
	PlaceMultipleOrders(ctx context.Context, req []okxrequests.PlaceOrder) (*okxresponses.PlaceOrder, error)
	ClosePosition(ctx context.Context, req okxrequests.ClosePosition) (*okxresponses.ClosePosition, error)
	GetOrderDetail(ctx context.Context, req okxrequests.OrderDetails) (*okxresponses.OrderList, error)
	GetOrderList(ctx context.Context, req okxrequests.OrderList) (*okxresponses.OrderList, error)
	GetOrderHistory(ctx context.Context, req okxrequests.OrderList, arch bool) (*okxresponses.OrderList, error)
	GetTransactionDetails(ctx context.Context, req okxrequests.TransactionDetails, arch bool) (*okxresponses.TransactionDetail, error)
	PlaceAlgoOrder(ctx context.Context, req okxrequests.PlaceAlgoOrder) (*okxresponses.PlaceAlgoOrder, error)
}

type Trade struct {
	client *Client
}

const (
	orderEndpoint                        = "/api/v5/trade/order"
	batchPlaceOrderEndpoint              = "/api/v5/trade/batch-orders"
	closePositionEndpoint                = "/api/v5/trade/close-position"
	getOrderDetainsEndpoint              = "/api/v5/trade/order" // TODO: fix this, since its the same route as POST order
	getOrderListEndpoint                 = "/api/v5/trade/orders-pending"
	placeAlgoOrderEnpdint                = "/api/v5/trade/order-algo"
	getTransactionDetailsEndpoint        = "/api/v5/trade/fills"
	getTransactionDetailsHistoryEndpoint = "/api/trade/fills-history"
	getOrdersHistoryEndpoint             = "/api/v5/trade/orders-history"
	getOrdersHistoryArchive              = "/api/v5/trade/orders-history-archive"
)

func NewTrade(clientOpt *ClientOptions, store limiter.Store, opts ...ClientOption) *Trade {
	trade := &Trade{
		client: NewClient(clientOpt, store, opts...),
	}
	trade.initLimiters()
	return trade
}

func (o *Trade) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		orderEndpoint:                        limiter.New(o.client.store, limiter.Rate{Limit: 60, Period: 2 * time.Second}),
		batchPlaceOrderEndpoint:              limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		getOrderListEndpoint:                 limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		placeAlgoOrderEnpdint:                limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		getTransactionDetailsEndpoint:        limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		getTransactionDetailsHistoryEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		getOrdersHistoryEndpoint:             limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		getOrdersHistoryArchive:              limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
	}
}

func (o *Trade) PlaceOrder(ctx context.Context, req []okxrequests.PlaceOrder) (*okxresponses.PlaceOrder, error) {
	response := &okxresponses.PlaceOrder{}
	p := orderEndpoint
	var tmp interface{}
	tmp = req[0]
	if len(req) > 1 {
		tmp = req
		p = batchPlaceOrderEndpoint
	}
	m := S2M(tmp)
	err := o.client.Do(ctx, http.MethodPost, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Trade) PlaceMultipleOrders(ctx context.Context, req []okxrequests.PlaceOrder) (*okxresponses.PlaceOrder, error) {
	response := &okxresponses.PlaceOrder{}
	p := batchPlaceOrderEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodPost, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Trade) ClosePosition(ctx context.Context, req okxrequests.ClosePosition) (*okxresponses.ClosePosition, error) {
	response := &okxresponses.ClosePosition{}
	p := closePositionEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodPost, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Trade) GetOrderDetail(ctx context.Context, req okxrequests.OrderDetails) (*okxresponses.OrderList, error) {
	response := &okxresponses.OrderList{}
	p := orderEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Trade) GetOrderList(ctx context.Context, req okxrequests.OrderList) (*okxresponses.OrderList, error) {
	response := &okxresponses.OrderList{}
	p := getOrderListEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Trade) GetOrderHistory(ctx context.Context, req okxrequests.OrderList, arch bool) (*okxresponses.OrderList, error) {
	response := &okxresponses.OrderList{}
	p := getOrdersHistoryEndpoint
	if arch {
		p = getOrdersHistoryArchive
	}
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Trade) GetTransactionDetails(ctx context.Context, req okxrequests.TransactionDetails, arch bool) (*okxresponses.TransactionDetail, error) {
	response := &okxresponses.TransactionDetail{}
	p := getTransactionDetailsEndpoint
	if arch {
		p = getTransactionDetailsHistoryEndpoint
	}
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Trade) PlaceAlgoOrder(ctx context.Context, req okxrequests.PlaceAlgoOrder) (*okxresponses.PlaceAlgoOrder, error) {
	response := &okxresponses.PlaceAlgoOrder{}
	p := placeAlgoOrderEnpdint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodPost, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
