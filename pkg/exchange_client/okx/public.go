package okx

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	okxrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/requests"
	okxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/responses"
)

const (
	getSystemTimeEndpoint  = "/api/v5/public/time"
	getInstrumentsEndpoint = "/api/v5/public/instruments"
)

type IOKXPublic interface {
	GetSystemTime(ctx context.Context) (*okxresponses.GetSystemTime, error)
	GetInstruments(ctx context.Context, req okxrequests.GetInstruments) (*okxresponses.GetInstruments, error)
}

type PublicData struct {
	client *Client
}

func NewPublicData(clientOpt *ClientOptions, store limiter.Store, opts ...ClientOption) *PublicData {
	publicData := &PublicData{
		client: NewClient(clientOpt, store, opts...),
	}
	publicData.initLimiters()
	return publicData
}

func (c *PublicData) initLimiters() {
	c.client.limiters = map[string]*limiter.Limiter{
		getSystemTimeEndpoint:  limiter.New(c.client.store, limiter.Rate{Limit: 10, Period: 2 * time.Second}),
		getInstrumentsEndpoint: limiter.New(c.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
	}
}

func (c *PublicData) GetSystemTime(ctx context.Context) (*okxresponses.GetSystemTime, error) {
	response := &okxresponses.GetSystemTime{}
	p := getSystemTimeEndpoint
	err := c.client.Do(ctx, http.MethodGet, p, false, response, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *PublicData) GetInstruments(ctx context.Context, req okxrequests.GetInstruments) (*okxresponses.GetInstruments, error) {
	response := &okxresponses.GetInstruments{}
	p := getInstrumentsEndpoint
	m := S2M(req)
	err := c.client.Do(ctx, http.MethodGet, p, false, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
