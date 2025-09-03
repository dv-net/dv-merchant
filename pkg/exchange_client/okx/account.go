package okx

import (
	"context"
	"net/http"
	"strings"
	"time"

	okxrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/requests"
	okxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/responses"

	"github.com/ulule/limiter/v3"
)

const (
	accountBalanceEndpoint = "/api/v5/account/balance"
	maxWithdrawalEndpoint  = "/api/v5/account/max-withdrawal"
	accountRiskEndpoint    = "/api/v5/account/account-position-risk"
)

type IOKXAccount interface {
	GetBalance(ctx context.Context, req okxrequests.GetAccountBalance) (*okxresponses.GetAccountBalance, error)
	GetMaxWithdrawals(ctx context.Context, req okxrequests.GetMaxWithdrawal) (*okxresponses.GetMaxWithdrawals, error)
	GetAccountAndRisks(ctx context.Context, req okxrequests.GetAccountAndPositionRisk) (*okxresponses.GetAccountAndPositionRisk, error)
}

type Account struct {
	client *Client
}

func NewAccount(clientOpt *ClientOptions, store limiter.Store, opts ...ClientOption) *Account {
	account := &Account{
		client: NewClient(clientOpt, store, opts...),
	}
	account.initLimiters()
	return account
}

func (o *Account) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		accountBalanceEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 10, Period: 2 * time.Second}),
		maxWithdrawalEndpoint:  limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: time.Second}),
		accountRiskEndpoint:    limiter.New(o.client.store, limiter.Rate{Limit: 10, Period: 2 * time.Second}),
	}
}

func (o *Account) GetBalance(ctx context.Context, req okxrequests.GetAccountBalance) (*okxresponses.GetAccountBalance, error) {
	response := &okxresponses.GetAccountBalance{}
	p := accountBalanceEndpoint
	m := S2M(req)
	if len(req.Ccy) > 0 {
		m["ccy"] = strings.Join(req.Ccy, ",")
	}
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Account) GetMaxWithdrawals(ctx context.Context, req okxrequests.GetMaxWithdrawal) (*okxresponses.GetMaxWithdrawals, error) {
	response := &okxresponses.GetMaxWithdrawals{}
	p := maxWithdrawalEndpoint
	m := S2M(req)
	if len(req.Ccy) > 0 {
		m["ccy"] = strings.Join(req.Ccy, ",")
	}
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (o *Account) GetAccountAndRisks(ctx context.Context, req okxrequests.GetAccountAndPositionRisk) (*okxresponses.GetAccountAndPositionRisk, error) {
	response := &okxresponses.GetAccountAndPositionRisk{}
	p := accountRiskEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
