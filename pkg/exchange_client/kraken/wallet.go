package kraken

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/responses"
)

const (
	depositMethodsEndpoint    = "/0/private/DepositMethods"
	depositAddressesEndpoint  = "/0/private/DepositAddresses"
	withdrawMethodsEndpoint   = "/0/private/WithdrawMethods"
	withdrawAddressesEndpoint = "/0/private/WithdrawAddresses"
	withdrawInfoEndpoint      = "/0/private/WithdrawInfo"
	withdrawEndpoint          = "/0/private/Withdraw"
	withdrawStatusEndpoint    = "/0/private/WithdrawStatus"
)

var ErrInvalidWalletRequest = errors.New("kraken: invalid wallet request")

type IKrakenWallet interface {
	GetDepositMethods(ctx context.Context, req *requests.DepositMethodsRequest) (responses.DepositMethodsResponse, error)
	GetDepositAddresses(ctx context.Context, req *requests.DepositAddressesRequest) (responses.DepositAddressesResponse, error)
	GetWithdrawMethods(ctx context.Context, req *requests.WithdrawMethodsRequest) (responses.WithdrawMethodsResponse, error)
	GetWithdrawAddresses(ctx context.Context, req *requests.WithdrawAddressesRequest) (responses.WithdrawAddressesResponse, error)
	GetWithdrawInfo(ctx context.Context, req *requests.WithdrawInfoRequest) (*responses.WithdrawInfoResponse, error)
	Withdraw(ctx context.Context, req *requests.WithdrawRequest) (*responses.WithdrawResponse, error)
	GetWithdrawStatus(ctx context.Context, req *requests.WithdrawStatusRequest) (responses.WithdrawStatusResponse, error)
}

var _ IKrakenWallet = (*WalletClient)(nil)

type WalletClient struct {
	client *Client
}

func NewWalletClient(client *Client) *WalletClient {
	return &WalletClient{client: client}
}

func (o *WalletClient) GetDepositMethods(ctx context.Context, req *requests.DepositMethodsRequest) (responses.DepositMethodsResponse, error) {
	if req == nil || req.Asset == "" {
		return nil, invalidWalletRequest("asset is required")
	}

	params := url.Values{"asset": {req.Asset}}
	setIfNotEmpty(params, "aclass", req.AClass)

	return DoPrivate[responses.DepositMethodsResponse](ctx, o.client, depositMethodsEndpoint, params)
}

func (o *WalletClient) GetDepositAddresses(ctx context.Context, req *requests.DepositAddressesRequest) (responses.DepositAddressesResponse, error) {
	if req == nil || req.Asset == "" || req.Method == "" {
		return nil, invalidWalletRequest("asset and method are required")
	}

	params := url.Values{"asset": {req.Asset}, "method": {req.Method}}
	if req.New {
		params.Set("new", "true")
	}
	setIfNotEmpty(params, "aclass", req.AClass)

	return DoPrivate[responses.DepositAddressesResponse](ctx, o.client, depositAddressesEndpoint, params)
}

func (o *WalletClient) GetWithdrawMethods(ctx context.Context, req *requests.WithdrawMethodsRequest) (responses.WithdrawMethodsResponse, error) {
	params := url.Values{}
	if req != nil {
		setIfNotEmpty(params, "asset", req.Asset)
		setIfNotEmpty(params, "aclass", req.AClass)
		setIfNotEmpty(params, "network", req.Network)
	}

	return DoPrivate[responses.WithdrawMethodsResponse](ctx, o.client, withdrawMethodsEndpoint, params)
}

func (o *WalletClient) GetWithdrawAddresses(ctx context.Context, req *requests.WithdrawAddressesRequest) (responses.WithdrawAddressesResponse, error) {
	params := url.Values{}
	if req != nil {
		setIfNotEmpty(params, "asset", req.Asset)
		setIfNotEmpty(params, "aclass", req.AClass)
		setIfNotEmpty(params, "method", req.Method)
		setIfNotEmpty(params, "key", req.Key)
		if req.Verified != nil {
			params.Set("verified", strconv.FormatBool(*req.Verified))
		}
	}

	return DoPrivate[responses.WithdrawAddressesResponse](ctx, o.client, withdrawAddressesEndpoint, params)
}

func (o *WalletClient) GetWithdrawInfo(ctx context.Context, req *requests.WithdrawInfoRequest) (*responses.WithdrawInfoResponse, error) {
	if req == nil || req.Asset == "" || req.Key == "" {
		return nil, invalidWalletRequest("asset and key are required")
	}
	if !req.Amount.IsPositive() {
		return nil, invalidWalletRequest("amount must be positive, got " + req.Amount.String())
	}

	params := url.Values{
		"asset":  {req.Asset},
		"key":    {req.Key},
		"amount": {req.Amount.String()},
	}

	res, err := DoPrivate[responses.WithdrawInfoResponse](ctx, o.client, withdrawInfoEndpoint, params)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (o *WalletClient) Withdraw(ctx context.Context, req *requests.WithdrawRequest) (*responses.WithdrawResponse, error) {
	if req == nil || req.Asset == "" || req.Key == "" {
		return nil, invalidWalletRequest("asset and key are required")
	}
	if !req.Amount.IsPositive() {
		return nil, invalidWalletRequest("amount must be positive, got " + req.Amount.String())
	}
	if req.MaxFee.IsNegative() {
		return nil, invalidWalletRequest("max fee must not be negative, got " + req.MaxFee.String())
	}

	params := url.Values{
		"asset":  {req.Asset},
		"key":    {req.Key},
		"amount": {req.Amount.String()},
	}
	setIfNotEmpty(params, "address", req.Address)
	setIfNotEmpty(params, "aclass", req.AClass)
	if req.MaxFee.IsPositive() {
		params.Set("max_fee", req.MaxFee.String())
	}

	res, err := DoPrivate[responses.WithdrawResponse](ctx, o.client, withdrawEndpoint, params)
	if err != nil {
		return nil, err
	}
	if res.RefID == "" {
		return nil, fmt.Errorf("kraken: withdrawal accepted without refid")
	}
	return &res, nil
}

func (o *WalletClient) GetWithdrawStatus(ctx context.Context, req *requests.WithdrawStatusRequest) (responses.WithdrawStatusResponse, error) {
	params := url.Values{}
	if req != nil {
		setIfNotEmpty(params, "asset", req.Asset)
		setIfNotEmpty(params, "aclass", req.AClass)
		setIfNotEmpty(params, "method", req.Method)
		if !req.Start.IsZero() {
			params.Set("start", strconv.FormatInt(req.Start.Unix(), 10))
		}
		if !req.End.IsZero() {
			params.Set("end", strconv.FormatInt(req.End.Unix(), 10))
		}
	}

	return DoPrivate[responses.WithdrawStatusResponse](ctx, o.client, withdrawStatusEndpoint, params)
}

func invalidWalletRequest(msg string) error {
	return fmt.Errorf("%w: %s", ErrInvalidWalletRequest, msg)
}

func setIfNotEmpty(params url.Values, key, value string) {
	if value != "" {
		params.Set(key, value)
	}
}
