package kucoin

import (
	"context"
	"net/http"
	"strings"
	"time"

	kucoinrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/requests"
	kucoinresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/responses"

	"github.com/ulule/limiter/v3"
)

const (
	getCurrencyList = "/api/v3/currencies"
	getCurrency     = "/api/v3/currencies/%s"
)

var _ = IKucoinMarket((*Market)(nil))

type IKucoinMarket interface {
	GetCurrencyList(ctx context.Context, req kucoinrequests.GetCurrencyList) (*kucoinresponses.GetCurrencyList, error)
	GetCurrency(ctx context.Context, req kucoinrequests.GetCurrency) (*kucoinresponses.GetCurrency, error)
}

type Market struct {
	client *Client
}

func NewMarket(clientOpt *ClientOptions, store limiter.Store, opts ...ClientOption) *Market {
	market := &Market{
		client: NewClient(clientOpt, store, opts...),
	}
	market.initLimiters()
	return market
}

func (o *Market) initLimiters() {
	// KuCoin market endpoints: These appear to be private endpoints requiring authentication
	// Conservative approach: 80 requests per 30 seconds for market data operations
	marketRate := limiter.Rate{Limit: 80, Period: 30 * time.Second}

	o.client.limiters = map[string]*limiter.Limiter{
		getCurrencyList: limiter.New(o.client.store, marketRate),
		getCurrency:     limiter.New(o.client.store, marketRate),
	}
}

func (o *Market) GetCurrencyList(ctx context.Context, req kucoinrequests.GetCurrencyList) (*kucoinresponses.GetCurrencyList, error) {
	response := &kucoinresponses.GetCurrencyList{}
	p := getCurrencyList
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Market) GetCurrency(ctx context.Context, req kucoinrequests.GetCurrency) (*kucoinresponses.GetCurrency, error) {
	response := &kucoinresponses.GetCurrency{}
	p := getCurrency
	m := S2M(req)
	if req.Currency != "" {
		p = strings.Replace(p, "%s", req.Currency, 1)
	}
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
