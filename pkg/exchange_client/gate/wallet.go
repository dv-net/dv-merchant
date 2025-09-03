package gateio

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"
)

const (
	getListChainsSupportedForSpecificCurrencyEndpoint = "/api/v4/wallet/currency_chains"
	getDepositAddressEndpoint                         = "/api/v4/wallet/deposit_address"
	getWithdrawalHistoryEndpoint                      = "/api/v4/wallet/withdrawals"
	getWithdrawalRulesEndpoint                        = "/api/v4/wallet/withdraw_status"
	createWithdrawalEndpoint                          = "/api/v4/withdrawals"
	getWalletSavedAddress                             = "/api/v4/wallet/saved_address"
)

var _ IGateWallet = (*WalletClient)(nil)

type IGateWallet interface {
	GetCurrencyChainsSupported(context.Context, *GetCurrencyChainsSupportedRequest) (*GetCurrencySupportedChainResponse, error)
	GetDepositAddress(ctx context.Context, dto *GetDepositAddressRequest) (*GetDepositAddressResponse, error)
	GetWithdrawalHistory(ctx context.Context, dto *GetWithdrawalHistoryRequest) (*GetWithdrawalHistoryResponse, error)
	GetWithdrawalRules(ctx context.Context, dto *GetWithdrawalRulesRequest) (*GetWithdrawalRulesResponse, error)
	CreateWithdrawal(ctx context.Context, dto *CreateWithdrawalRequest) (*CreateWithdrawalResponse, error)
	GetWalletSavedAddresses(ctx context.Context, dto *GetWalletSavedAddressesRequest) (*GetWalletSavedAddressesResponse, error)
}

type WalletClient struct {
	client *Client
}

func NewWalletClient(opt *ClientOptions, store limiter.Store) *WalletClient {
	wallet := &WalletClient{
		client: NewClient(opt, store),
	}
	wallet.initLimiters()
	return wallet
}

func (o *WalletClient) initLimiters() {
	o.client.limiters = map[string]*limiter.Limiter{
		getListChainsSupportedForSpecificCurrencyEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 20, Period: 2 * time.Second}),
	}
}

func (o *WalletClient) GetCurrencyChainsSupported(ctx context.Context, dto *GetCurrencyChainsSupportedRequest) (*GetCurrencySupportedChainResponse, error) {
	p := getListChainsSupportedForSpecificCurrencyEndpoint
	response := &GetCurrencySupportedChainResponse{}
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response.Data, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetDepositAddress(ctx context.Context, dto *GetDepositAddressRequest) (*GetDepositAddressResponse, error) {
	p := getDepositAddressEndpoint
	response := &GetDepositAddressResponse{}
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response.Data, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetWithdrawalHistory(ctx context.Context, dto *GetWithdrawalHistoryRequest) (*GetWithdrawalHistoryResponse, error) {
	p := getWithdrawalHistoryEndpoint
	response := &GetWithdrawalHistoryResponse{}
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response.Data, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetWithdrawalRules(ctx context.Context, dto *GetWithdrawalRulesRequest) (*GetWithdrawalRulesResponse, error) {
	p := getWithdrawalRulesEndpoint
	response := &GetWithdrawalRulesResponse{}
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response.Data, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) CreateWithdrawal(ctx context.Context, dto *CreateWithdrawalRequest) (*CreateWithdrawalResponse, error) {
	p := createWithdrawalEndpoint
	response := &CreateWithdrawalResponse{}
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodPost, p, true, &response.Data, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetWalletSavedAddresses(ctx context.Context, dto *GetWalletSavedAddressesRequest) (*GetWalletSavedAddressesResponse, error) {
	p := getWalletSavedAddress
	response := &GetWalletSavedAddressesResponse{}
	m := S2M(dto)
	err := o.client.Do(ctx, http.MethodGet, p, true, &response.Data, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
