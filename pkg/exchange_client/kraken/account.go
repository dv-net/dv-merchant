package kraken

import (
	"context"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/responses"
)

const (
	getBalanceEndpoint   = "/0/private/Balance"
	getBalanceExEndpoint = "/0/private/BalanceEx"
)

type IKrakenAccount interface {
	GetBalance(ctx context.Context) (responses.BalanceResponse, error)
	GetBalanceEx(ctx context.Context) (responses.BalanceExResponse, error)
}

var _ IKrakenAccount = (*AccountClient)(nil)

type AccountClient struct {
	client *Client
}

func NewAccountClient(client *Client) *AccountClient {
	return &AccountClient{client: client}
}

func (o *AccountClient) GetBalance(ctx context.Context) (responses.BalanceResponse, error) {
	return DoPrivate[responses.BalanceResponse](ctx, o.client, getBalanceEndpoint, nil)
}

func (o *AccountClient) GetBalanceEx(ctx context.Context) (responses.BalanceExResponse, error) {
	return DoPrivate[responses.BalanceExResponse](ctx, o.client, getBalanceExEndpoint, nil)
}
