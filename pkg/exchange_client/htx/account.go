package htx

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	htxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/models"
	htxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/responses"
)

const (
	getAllAccountsEndpoint    = "/v1/account/accounts"
	getAccountBalanceEndpoint = "/v1/account/accounts/%d/balance"
)

type IHTXAccount interface {
	GetAllAccounts(ctx context.Context) (*htxresponses.GetAccounts, error)
	GetAccountBalance(ctx context.Context, account *htxmodels.Account) (*htxresponses.GetAccountBalance, error)
}

func NewAccountClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *AccountClient {
	account := &AccountClient{
		client: NewClient(opt, store, opts...),
	}
	account.initLimiters()
	return account
}

type AccountClient struct {
	client *Client
}

func (o *AccountClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getAllAccountsEndpoint:    limiter.New(o.client.store, limiter.Rate{Limit: 30, Period: 3 * time.Second}),
		getAccountBalanceEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 30, Period: 3 * time.Second}),
	}
}

func (o *AccountClient) GetAllAccounts(ctx context.Context) (*htxresponses.GetAccounts, error) {
	response := &htxresponses.GetAccounts{}
	p := getAllAccountsEndpoint
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *AccountClient) GetAccountBalance(ctx context.Context, account *htxmodels.Account) (*htxresponses.GetAccountBalance, error) {
	response := &htxresponses.GetAccountBalance{}
	p := fmt.Sprintf(getAccountBalanceEndpoint, account.ID)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}
