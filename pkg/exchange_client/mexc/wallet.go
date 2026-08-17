package mexc

import (
	"context"
	"net/http"
	"time"

	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/mexc/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/mexc/responses"
)

const (
	getCoinsConfigEndpoint  = "/api/v3/capital/config/getall"
	depositAddressEndpoint  = "/api/v3/capital/deposit/address"
	withdrawEndpoint        = "/api/v3/capital/withdraw"
	withdrawHistoryEndpoint = "/api/v3/capital/withdraw/history"

	defaultWithdrawHistoryLimit = "1000"
)

type IMexcWallet interface {
	GetCoinsConfig(ctx context.Context) (responses.GetCoinsConfigResponse, error)
	GetDepositAddress(ctx context.Context, req *requests.GetDepositAddressRequest) (responses.GetDepositAddressResponse, error)
	CreateDepositAddress(ctx context.Context, req *requests.CreateDepositAddressRequest) (*responses.CreateDepositAddressResponse, error)
	Withdraw(ctx context.Context, req *requests.WithdrawRequest) (*responses.WithdrawResponse, error)
	GetWithdrawHistory(ctx context.Context, req *requests.GetWithdrawHistoryRequest) (responses.GetWithdrawHistoryResponse, error)
}

var _ IMexcWallet = (*WalletClient)(nil)

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
		getCoinsConfigEndpoint:  limiter.New(o.client.store, limiter.Rate{Limit: 2, Period: time.Second}),
		depositAddressEndpoint:  limiter.New(o.client.store, limiter.Rate{Limit: 2, Period: time.Second}),
		withdrawEndpoint:        limiter.New(o.client.store, limiter.Rate{Limit: 1, Period: time.Second}),
		withdrawHistoryEndpoint: limiter.New(o.client.store, limiter.Rate{Limit: 2, Period: time.Second}),
	}
}

func (o *WalletClient) GetCoinsConfig(ctx context.Context) (responses.GetCoinsConfigResponse, error) {
	res := responses.GetCoinsConfigResponse{}
	if err := o.client.Do(ctx, http.MethodGet, getCoinsConfigEndpoint, true, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (o *WalletClient) GetDepositAddress(ctx context.Context, req *requests.GetDepositAddressRequest) (responses.GetDepositAddressResponse, error) {
	params := map[string]string{"coin": req.Coin}
	if req.Network != "" {
		params["network"] = req.Network
	}

	res := responses.GetDepositAddressResponse{}
	if err := o.client.Do(ctx, http.MethodGet, depositAddressEndpoint, true, &res, params); err != nil {
		return nil, err
	}
	return res, nil
}

func (o *WalletClient) CreateDepositAddress(ctx context.Context, req *requests.CreateDepositAddressRequest) (*responses.CreateDepositAddressResponse, error) {
	params := map[string]string{"coin": req.Coin, "network": req.Network}

	res := &responses.CreateDepositAddressResponse{}
	if err := o.client.Do(ctx, http.MethodPost, depositAddressEndpoint, true, res, params); err != nil {
		return nil, err
	}
	return res, nil
}

func (o *WalletClient) Withdraw(ctx context.Context, req *requests.WithdrawRequest) (*responses.WithdrawResponse, error) {
	params := map[string]string{
		"coin":    req.Coin,
		"address": req.Address,
		"amount":  req.Amount,
		"netWork": req.Network,
	}
	if req.ContractAddress != "" {
		params["contractAddress"] = req.ContractAddress
	}
	if req.Memo != "" {
		params["memo"] = req.Memo
	}

	res := &responses.WithdrawResponse{}
	if err := o.client.Do(ctx, http.MethodPost, withdrawEndpoint, true, res, params); err != nil {
		return nil, err
	}
	return res, nil
}

func (o *WalletClient) GetWithdrawHistory(ctx context.Context, req *requests.GetWithdrawHistoryRequest) (responses.GetWithdrawHistoryResponse, error) {
	params := map[string]string{"limit": defaultWithdrawHistoryLimit}
	if req.Coin != "" {
		params["coin"] = req.Coin
	}

	res := responses.GetWithdrawHistoryResponse{}
	if err := o.client.Do(ctx, http.MethodGet, withdrawHistoryEndpoint, true, &res, params); err != nil {
		return nil, err
	}
	return res, nil
}
