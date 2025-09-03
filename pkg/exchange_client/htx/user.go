package htx

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	htxrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/requests"
	htxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/responses"
)

const (
	getAPIKeyInformationEndpoint = "/v2/user/api-key" //nolint:gosec
	getUserUIDEndpoint           = "/v2/user/uid"
)

type IHTXUser interface {
	GetAPIKeyInformation(ctx context.Context, dto *htxrequests.GetAPIKeyInformationRequest) (*htxresponses.GetAPIKeyInformationResponse, error)
	GetUserUID(ctx context.Context) (*htxresponses.GetUserUID, error)
}

type UserClient struct {
	client *Client
}

func NewUserClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *UserClient {
	user := &UserClient{
		client: NewClient(opt, store, opts...),
	}
	user.initLimiters()
	return user
}

func (o *UserClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getAPIKeyInformationEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 30, Period: 3 * time.Second}),
		getUserUIDEndpoint:           limiter.New(o.client.store, limiter.Rate{Limit: 2, Period: 3 * time.Second}),
	}
}

func (o *UserClient) GetAPIKeyInformation(ctx context.Context, dto *htxrequests.GetAPIKeyInformationRequest) (*htxresponses.GetAPIKeyInformationResponse, error) {
	response := &htxresponses.GetAPIKeyInformationResponse{}
	p := getAPIKeyInformationEndpoint
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *UserClient) GetUserUID(ctx context.Context) (*htxresponses.GetUserUID, error) {
	response := &htxresponses.GetUserUID{}
	p := getUserUIDEndpoint
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}
