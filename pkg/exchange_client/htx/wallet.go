package htx

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	htxrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/requests"
	htxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/responses"
)

const (
	postVirtualCurrencyWithdrawalEndpoint = "/v1/dw/withdraw/api/create"
	getWithdrawalAddressEndpoint          = "/v2/account/withdraw/address"
	getDepositAddressEndpoint             = "/v2/account/deposit/address"
	postCancelWithdrawalEndpoint          = "/v1/dw/withdraw-virtual/{withdraw-id}/cancel"
	getWithdrawalByClientIDEndpoint       = "/v1/query/withdraw/client-order-id"
	getWithdrawalDepositHistoryEndpoint   = "/v1/query/deposit-withdraw"
)

type IHTXWallet interface {
	WithdrawVirtualCurrency(ctx context.Context, dto *htxrequests.WithdrawVirtualCurrencyRequest) (*htxresponses.WithdrawalVirtual, error)
	GetWithdrawalAddress(ctx context.Context, dto *htxrequests.WithdrawalAddressRequest) (*htxresponses.GetWithdrawalAddress, error)
	GetDepositAddress(ctx context.Context, dto *htxrequests.DepositAddressRequest) (*htxresponses.GetDepositAddress, error)
	CancelWithdrawal(ctx context.Context, dto *htxrequests.CancelWithdrawalRequest) (*htxresponses.CancelWithdrawal, error)
	GetWithdrawalByClientID(ctx context.Context, dto *htxrequests.WithdrawalByClientIDRequest) (*htxresponses.GetWithdrawalByClientID, error)
	GetWithdrawalDepositHistory(ctx context.Context, dto *htxrequests.WithdrawalDepositHistoryRequest) (*htxresponses.GetWithdrawalDepositHistory, error)
}

type WalletClient struct {
	client *Client
}

func NewWalletClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *WalletClient {
	wallet := &WalletClient{
		client: NewClient(opt, store, opts...),
	}
	wallet.initLimiters()
	return wallet
}

func (o *WalletClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		postVirtualCurrencyWithdrawalEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 5, Period: 2 * time.Second}),
		getWithdrawalAddressEndpoint:          limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		getDepositAddressEndpoint:             limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		postCancelWithdrawalEndpoint:          limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		getWithdrawalByClientIDEndpoint:       limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
		getWithdrawalDepositHistoryEndpoint:   limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
	}
}

func (o *WalletClient) WithdrawVirtualCurrency(ctx context.Context, dto *htxrequests.WithdrawVirtualCurrencyRequest) (*htxresponses.WithdrawalVirtual, error) {
	response := &htxresponses.WithdrawalVirtual{}
	p := postVirtualCurrencyWithdrawalEndpoint
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodPost, p, true, &response, m)
	if err != nil {
		return nil, err
	}
	if response.WithdrawalTransferID == 0 {
		return nil, errors.New("withdrawal failed")
	}
	return response, nil
}

func (o *WalletClient) GetWithdrawalAddress(ctx context.Context, dto *htxrequests.WithdrawalAddressRequest) (*htxresponses.GetWithdrawalAddress, error) {
	response := &htxresponses.GetWithdrawalAddress{}
	p := getWithdrawalAddressEndpoint
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetDepositAddress(ctx context.Context, dto *htxrequests.DepositAddressRequest) (*htxresponses.GetDepositAddress, error) {
	response := &htxresponses.GetDepositAddress{}
	p := getDepositAddressEndpoint
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetWithdrawalByClientID(ctx context.Context, dto *htxrequests.WithdrawalByClientIDRequest) (*htxresponses.GetWithdrawalByClientID, error) {
	response := &htxresponses.GetWithdrawalByClientID{}
	p := getWithdrawalByClientIDEndpoint
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetWithdrawalDepositHistory(ctx context.Context, dto *htxrequests.WithdrawalDepositHistoryRequest) (*htxresponses.GetWithdrawalDepositHistory, error) {
	response := &htxresponses.GetWithdrawalDepositHistory{}
	p := getWithdrawalDepositHistoryEndpoint
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) CancelWithdrawal(ctx context.Context, dto *htxrequests.CancelWithdrawalRequest) (*htxresponses.CancelWithdrawal, error) {
	response := &htxresponses.CancelWithdrawal{}
	p := fmt.Sprint(postCancelWithdrawalEndpoint, dto.WithdrawalTransferID)
	err := o.client.Do(ctx, http.MethodPost, p, true, response, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}
