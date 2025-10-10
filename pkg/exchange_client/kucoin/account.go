package kucoin

import (
	"context"
	"net/http"
	"time"

	kucoinrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/requests"
	kucoinresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/responses"

	"github.com/ulule/limiter/v3"
)

const (
	getAPIKeyInfo        = "/api/v1/user/api-key" //nolint:gosec
	getAccountList       = "/api/v1/accounts"
	getDepositAddress    = "/api/v3/deposit-addresses"
	createDepositAddress = "/api/v3/deposit-address/create"
	getWithdrawalHistory = "/api/v1/withdrawals"
	createWithdrawal     = "/api/v3/withdrawals"
	createFlexTransfer   = "/api/v3/accounts/universal-transfer"
)

var _ IKucoinAccount = (*Account)(nil)

type IKucoinAccount interface {
	GetAPIKeyInfo(ctx context.Context, req kucoinrequests.GetAPIKeyInfo) (*kucoinresponses.GetAPIKeyInfo, error)
	GetAccountList(ctx context.Context, req kucoinrequests.GetAccountList) (*kucoinresponses.GetAccountList, error)
	GetDepositAddress(ctx context.Context, req kucoinrequests.GetDepositAddress) (*kucoinresponses.GetDepositAddress, error)
	CreateDepositAddress(ctx context.Context, req kucoinrequests.CreateDepositAddress) (*kucoinresponses.CreateDepositAddress, error)
	CreateWithdrawal(ctx context.Context, req kucoinrequests.CreateWithdrawal) (*kucoinresponses.CreateWithdrawal, error)
	CreateFlexTransfer(ctx context.Context, req kucoinrequests.FlexTransfer) (*kucoinresponses.FlexTransfer, error)
	GetWithdrawalHistory(ctx context.Context, req kucoinrequests.GetWithdrawalHistory) (*kucoinresponses.GetWithdrawalHistory, error)
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
	// KuCoin private endpoints: VIP0 allows 4000 requests per 30 seconds for Management pool
	// Conservative approach: 50 requests per 30 seconds for account operations
	accountRate := limiter.Rate{Limit: 50, Period: 30 * time.Second}
	withdrawalRate := limiter.Rate{Limit: 5, Period: 30 * time.Second} // TODO: Replace when determine actual rate-limits

	o.client.limiters = map[string]*limiter.Limiter{
		getAPIKeyInfo:        limiter.New(o.client.store, accountRate),
		getAccountList:       limiter.New(o.client.store, accountRate),
		getDepositAddress:    limiter.New(o.client.store, accountRate),
		createDepositAddress: limiter.New(o.client.store, accountRate),
		getWithdrawalHistory: limiter.New(o.client.store, accountRate),
		createWithdrawal:     limiter.New(o.client.store, withdrawalRate),
		createFlexTransfer:   limiter.New(o.client.store, accountRate),
	}
}

// api-permission: [kucoinmodels.PermissionGeneral - "General"]
func (o *Account) GetAPIKeyInfo(ctx context.Context, req kucoinrequests.GetAPIKeyInfo) (*kucoinresponses.GetAPIKeyInfo, error) {
	response := &kucoinresponses.GetAPIKeyInfo{}
	p := getAPIKeyInfo
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// api-permission: [kucoinmodels.PermissionGeneral - "General"]
func (o *Account) GetAccountList(ctx context.Context, req kucoinrequests.GetAccountList) (*kucoinresponses.GetAccountList, error) {
	response := &kucoinresponses.GetAccountList{}
	p := getAccountList
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// api-permission: [kucoinmodels.PermissionGeneral - "General"]
func (o *Account) GetDepositAddress(ctx context.Context, req kucoinrequests.GetDepositAddress) (*kucoinresponses.GetDepositAddress, error) {
	response := &kucoinresponses.GetDepositAddress{}
	p := getDepositAddress
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// api-permission: [kucoinmodels.PermissionGeneral - "General"]
func (o *Account) CreateDepositAddress(ctx context.Context, req kucoinrequests.CreateDepositAddress) (*kucoinresponses.CreateDepositAddress, error) {
	response := &kucoinresponses.CreateDepositAddress{}
	p := createDepositAddress
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodPost, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *Account) CreateWithdrawal(ctx context.Context, req kucoinrequests.CreateWithdrawal) (*kucoinresponses.CreateWithdrawal, error) {
	response := &kucoinresponses.CreateWithdrawal{}
	p := createWithdrawal
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodPost, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// api-permission: [kucoinmodels.PermissionGeneral - "General"]
func (o *Account) GetWithdrawalHistory(ctx context.Context, req kucoinrequests.GetWithdrawalHistory) (*kucoinresponses.GetWithdrawalHistory, error) {
	response := &kucoinresponses.GetWithdrawalHistory{}
	p := getWithdrawalHistory
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// api-permission: [kucoinmodels.PermissionTransfer - "Transfer"]
func (o *Account) CreateFlexTransfer(ctx context.Context, req kucoinrequests.FlexTransfer) (*kucoinresponses.FlexTransfer, error) {
	response := &kucoinresponses.FlexTransfer{}
	p := createFlexTransfer
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodPost, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
