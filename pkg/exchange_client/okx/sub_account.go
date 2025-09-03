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
	subAccountTransferEndpoint        = "/api/v5/asset/subaccount/transfer"
	subAccountBalanceEndpoint         = "/api/v5/account/subaccount/balances"
	subAccountHistoryTransferEndpoint = "/api/v5/asset/subaccount/bills"
)

type IOKXSubAccount interface {
	GetBalance(ctx context.Context, req okxrequests.GetBalance) (*okxresponses.GetSubAccountBalance, error)
	HistoryTransfer(ctx context.Context, req okxrequests.HistoryTransfer) (*okxresponses.HistoryTransfer, error)
	ManageTransfers(ctx context.Context, req okxrequests.ManageTransfers) (*okxresponses.ManageTransfer, error)
}

type SubAccount struct {
	client *Client
}

func NewSubAccount(clientOpt *ClientOptions, store limiter.Store, opts ...ClientOption) *SubAccount {
	subAccount := &SubAccount{
		client: NewClient(clientOpt, store, opts...),
	}
	subAccount.initLimiters()
	return subAccount
}

func (o *SubAccount) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		subAccountTransferEndpoint:        limiter.New(o.client.store, limiter.Rate{Limit: 1, Period: time.Second}),
		subAccountBalanceEndpoint:         limiter.New(o.client.store, limiter.Rate{Limit: 6, Period: 2 * time.Second}),
		subAccountHistoryTransferEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 6, Period: time.Second}),
	}
}

func (o *SubAccount) GetBalance(ctx context.Context, req okxrequests.GetBalance) (*okxresponses.GetSubAccountBalance, error) {
	response := &okxresponses.GetSubAccountBalance{}
	p := subAccountBalanceEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *SubAccount) HistoryTransfer(ctx context.Context, req okxrequests.HistoryTransfer) (*okxresponses.HistoryTransfer, error) {
	response := &okxresponses.HistoryTransfer{}
	p := subAccountHistoryTransferEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *SubAccount) ManageTransfers(ctx context.Context, req okxrequests.ManageTransfers) (*okxresponses.ManageTransfer, error) {
	response := &okxresponses.ManageTransfer{}
	p := subAccountTransferEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
