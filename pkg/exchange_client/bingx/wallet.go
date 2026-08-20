package bingx

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bingx/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bingx/responses"
)

const (
	getCoinsConfigEndpoint  = "/openApi/wallets/v1/capital/config/getall"
	depositAddressEndpoint  = "/openApi/wallets/v1/capital/deposit/address"
	withdrawEndpoint        = "/openApi/wallets/v1/capital/withdraw/apply"
	withdrawHistoryEndpoint = "/openApi/api/v3/capital/withdraw/history"
)

const (
	WalletTypeFund        = 1
	WalletTypeStdFutures  = 2
	WalletTypePerpFutures = 3

	defaultDepositAddressLimit  = "100"
	defaultWithdrawHistoryLimit = "1000"
)

type IBingxWallet interface {
	GetCoinsConfig(ctx context.Context, coin string) (responses.GetCoinsConfigResponse, error)
	GetDepositAddress(ctx context.Context, req *requests.GetDepositAddressRequest) (*responses.GetDepositAddressResponse, error)
	Withdraw(ctx context.Context, req *requests.WithdrawRequest) (*responses.WithdrawResponse, error)
	GetWithdrawHistory(ctx context.Context, req *requests.GetWithdrawHistoryRequest) (responses.GetWithdrawHistoryResponse, error)
}

var _ IBingxWallet = (*WalletClient)(nil)

type WalletClient struct {
	client *Client
}

const (
	// TODO: bingx has a strict rate limit, requests may still fail. Drop the dependency on this interval.
	depositAddressInterval = 1200 * time.Millisecond
)

func NewWalletClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) *WalletClient {
	wallet := &WalletClient{
		client: NewClient(opt, store, opts...),
	}
	wallet.initLimiters()
	return wallet
}

func (o *WalletClient) initLimiters() {
	o.client.intervals = map[string]time.Duration{
		getCoinsConfigEndpoint:  200 * time.Millisecond,
		depositAddressEndpoint:  depositAddressInterval,
		withdrawEndpoint:        500 * time.Millisecond,
		withdrawHistoryEndpoint: 200 * time.Millisecond,
	}
}

func (o *WalletClient) GetCoinsConfig(ctx context.Context, coin string) (responses.GetCoinsConfigResponse, error) {
	params := map[string]string{}
	if coin != "" {
		params["coin"] = coin
	}

	res := responses.GetCoinsConfigResponse{}
	if err := o.client.Do(ctx, http.MethodGet, getCoinsConfigEndpoint, true, &res, params); err != nil {
		return nil, err
	}
	return res, nil
}

func (o *WalletClient) GetDepositAddress(ctx context.Context, req *requests.GetDepositAddressRequest) (*responses.GetDepositAddressResponse, error) {
	params := map[string]string{
		"coin":   req.Coin,
		"limit":  defaultDepositAddressLimit,
		"offset": strconv.Itoa(req.Offset),
	}
	if req.Limit > 0 {
		params["limit"] = strconv.Itoa(req.Limit)
	}

	res := &responses.GetDepositAddressResponse{}
	if err := o.client.Do(ctx, http.MethodGet, depositAddressEndpoint, true, res, params); err != nil {
		return nil, err
	}
	return res, nil
}

func (o *WalletClient) Withdraw(ctx context.Context, req *requests.WithdrawRequest) (*responses.WithdrawResponse, error) {
	walletType := req.WalletType
	if walletType == 0 {
		walletType = WalletTypeFund
	}

	params := map[string]string{
		"coin":       req.Coin,
		"address":    req.Address,
		"amount":     req.Amount,
		"walletType": strconv.Itoa(walletType),
	}
	if req.Network != "" {
		params["network"] = req.Network
	}
	if req.AddressTag != "" {
		params["addressTag"] = req.AddressTag
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
