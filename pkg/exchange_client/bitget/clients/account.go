package clients

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget"
	bitgetrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/requests"
	bitgetresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/responses"
)

var _ IBitgetAccount = (*AccountClient)(nil)

const (
	getAccountAssetsEndpoint  = "/api/v2/spot/account/assets"
	getDepositAddressEndpoint = "/api/v2/spot/wallet/deposit-address"
	getDepositRecordsEndpoint = "/api/v2/spot/wallet/deposit-records"
	getWithdrawalRecords      = "/api/v2/spot/wallet/withdrawal-records"
	postWalletWithdrawal      = "/api/v2/spot/wallet/withdrawal"
)

type IBitgetAccount interface {
	AccountAssets(context.Context, *bitgetrequests.AccountAssetsRequest) (*bitgetresponses.AccountAssetsResponse, error)
	DepositAddress(context.Context, *bitgetrequests.DepositAddressRequest) (*bitgetresponses.DepositAddressResponse, error)
	DepositRecords(context.Context, *bitgetrequests.DepositRecordsRequest) (*bitgetresponses.DepositRecordsResponse, error)
	WithdrawalRecords(context.Context, *bitgetrequests.WithdrawalRecordsRequest) (*bitgetresponses.WithdrawalRecordsResponse, error)
	WalletWithdrawal(context.Context, *bitgetrequests.WalletWithdrawalRequest) (*bitgetresponses.WalletWithdrawalResponse, error)
}

func NewAccountClient(opt *ClientOptions, store limiter.Store, signer bitget.ISigner) *AccountClient {
	account := &AccountClient{
		client: NewClient(opt, store, signer),
	}
	account.initLimiters()
	return account
}

type AccountClient struct {
	client *Client
}

func (o *AccountClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getAccountAssetsEndpoint:  limiter.New(o.client.store, limiter.Rate{Limit: 10, Period: time.Second}),
		getDepositAddressEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 10, Period: time.Second}),
		getDepositRecordsEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 10, Period: time.Second}),
		getWithdrawalRecords:      limiter.New(o.client.store, limiter.Rate{Limit: 10, Period: time.Second}),
		postWalletWithdrawal:      limiter.New(o.client.store, limiter.Rate{Limit: 5, Period: time.Second}),
	}
}

func (o *AccountClient) AccountAssets(ctx context.Context, req *bitgetrequests.AccountAssetsRequest) (*bitgetresponses.AccountAssetsResponse, error) {
	response := &bitgetresponses.AccountAssetsResponse{}
	p := getAccountAssetsEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *AccountClient) DepositAddress(ctx context.Context, req *bitgetrequests.DepositAddressRequest) (*bitgetresponses.DepositAddressResponse, error) {
	response := &bitgetresponses.DepositAddressResponse{}
	p := getDepositAddressEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *AccountClient) DepositRecords(ctx context.Context, req *bitgetrequests.DepositRecordsRequest) (*bitgetresponses.DepositRecordsResponse, error) {
	response := &bitgetresponses.DepositRecordsResponse{}
	p := getDepositRecordsEndpoint
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *AccountClient) WithdrawalRecords(ctx context.Context, req *bitgetrequests.WithdrawalRecordsRequest) (*bitgetresponses.WithdrawalRecordsResponse, error) {
	response := &bitgetresponses.WithdrawalRecordsResponse{}
	p := getWithdrawalRecords
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *AccountClient) WalletWithdrawal(ctx context.Context, req *bitgetrequests.WalletWithdrawalRequest) (*bitgetresponses.WalletWithdrawalResponse, error) {
	response := &bitgetresponses.WalletWithdrawalResponse{}
	p := postWalletWithdrawal
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodPost, p, true, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
