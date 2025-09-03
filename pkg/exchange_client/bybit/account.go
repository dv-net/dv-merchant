package bybit

import (
	"context"
	"net/http"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/responses"

	"github.com/ulule/limiter/v3"
)

type IBybitAccount interface {
	GetTranscationLog(ctx context.Context, req *requests.GetTranscationLogRequest) (*responses.BaseResponse[responses.GetTransactionLogResponse], error)
	GetTradingBalance(ctx context.Context, req *requests.GetTradingBalanceRequest) (*responses.BaseResponse[responses.GetTradingBalanceResponse], error)
	GetAllCoinsBalance(ctx context.Context, req *requests.GetFundingBalanceRequest) (*responses.BaseResponse[responses.GetAllCoinsBalanceResponse], error)
	GetAPIKeyInfo(ctx context.Context) (*responses.BaseResponse[responses.GetAPIKeyInfoResponse], error)
	GetAccountInfo(ctx context.Context) (*responses.BaseResponse[responses.GetAccountInfoResponse], error)
	UpgradeToUnifiedTradingAccount(ctx context.Context) (*responses.BaseResponse[responses.UpgradeToUnifiedTradingAccountResponse], error)
	GetDepositAddress(ctx context.Context, req *requests.GetDepositAddressRequest) (*responses.BaseResponse[responses.GetDepositAddressResponse], error)
	CreateInternalTransfer(ctx context.Context, req *requests.CreateInternalTransferRequest) (*responses.BaseResponse[responses.CreateInternalTransferResponse], error)
	CreateWithdraw(ctx context.Context, req *requests.CreateWithdrawRequest) (*responses.BaseResponse[responses.CreateWithdrawResponse], error)
	GetWithdraw(ctx context.Context, req *requests.GetWithdrawRequest) (*responses.BaseResponse[responses.GetWithdrawResponse], error)
}

var _ IBybitAccount = (*AccountClient)(nil)

type AccountClient struct {
	client *Client
}

func NewAccount(opt *ClientOptions, store limiter.Store, opts ...ClientOption) IBybitAccount {
	client := NewClient(opt, store, opts...)
	return &AccountClient{
		client: client,
	}
}

func (o *AccountClient) GetTranscationLog(ctx context.Context, req *requests.GetTranscationLogRequest) (*responses.BaseResponse[responses.GetTransactionLogResponse], error) {
	endpoint := "/v5/account/transaction-log"

	params := S2M(req)
	var resp responses.BaseResponse[responses.GetTransactionLogResponse]
	err := o.client.Do(ctx, http.MethodGet, endpoint, true, &resp, params)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *AccountClient) GetTradingBalance(ctx context.Context, req *requests.GetTradingBalanceRequest) (*responses.BaseResponse[responses.GetTradingBalanceResponse], error) {
	endpoint := "/v5/account/wallet-balance"

	params := S2M(req)
	var resp responses.BaseResponse[responses.GetTradingBalanceResponse]
	err := o.client.Do(ctx, http.MethodGet, endpoint, true, &resp, params)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *AccountClient) GetAllCoinsBalance(ctx context.Context, req *requests.GetFundingBalanceRequest) (*responses.BaseResponse[responses.GetAllCoinsBalanceResponse], error) {
	endpoint := "/v5/asset/transfer/query-account-coins-balance"

	params := S2M(req)
	var resp responses.BaseResponse[responses.GetAllCoinsBalanceResponse]
	err := o.client.Do(ctx, http.MethodGet, endpoint, true, &resp, params)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *AccountClient) GetAPIKeyInfo(ctx context.Context) (*responses.BaseResponse[responses.GetAPIKeyInfoResponse], error) {
	endpoint := "/v5/user/query-api"

	var resp responses.BaseResponse[responses.GetAPIKeyInfoResponse]
	err := o.client.Do(ctx, http.MethodGet, endpoint, true, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *AccountClient) GetAccountInfo(ctx context.Context) (*responses.BaseResponse[responses.GetAccountInfoResponse], error) {
	endpoint := "/v5/account/info"

	var resp responses.BaseResponse[responses.GetAccountInfoResponse]
	err := o.client.Do(ctx, http.MethodGet, endpoint, true, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *AccountClient) UpgradeToUnifiedTradingAccount(ctx context.Context) (*responses.BaseResponse[responses.UpgradeToUnifiedTradingAccountResponse], error) {
	endpoint := "/v5/account/upgrade-to-uta"

	var resp responses.BaseResponse[responses.UpgradeToUnifiedTradingAccountResponse]
	err := o.client.Do(ctx, http.MethodPost, endpoint, true, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *AccountClient) GetDepositAddress(ctx context.Context, req *requests.GetDepositAddressRequest) (*responses.BaseResponse[responses.GetDepositAddressResponse], error) {
	endpoint := "/v5/asset/deposit/query-address"

	params := S2M(req)
	var resp responses.BaseResponse[responses.GetDepositAddressResponse]
	err := o.client.Do(ctx, http.MethodGet, endpoint, true, &resp, params)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *AccountClient) CreateInternalTransfer(ctx context.Context, req *requests.CreateInternalTransferRequest) (*responses.BaseResponse[responses.CreateInternalTransferResponse], error) {
	endpoint := "/v5/asset/transfer/inter-transfer"

	params := S2M(req)
	var resp responses.BaseResponse[responses.CreateInternalTransferResponse]
	err := o.client.Do(ctx, http.MethodPost, endpoint, true, &resp, params)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *AccountClient) CreateWithdraw(ctx context.Context, req *requests.CreateWithdrawRequest) (*responses.BaseResponse[responses.CreateWithdrawResponse], error) {
	endpoint := "/v5/asset/withdraw/create"

	params := S2M(req)
	var resp responses.BaseResponse[responses.CreateWithdrawResponse]
	err := o.client.Do(ctx, http.MethodPost, endpoint, true, &resp, params)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *AccountClient) GetWithdraw(ctx context.Context, req *requests.GetWithdrawRequest) (*responses.BaseResponse[responses.GetWithdrawResponse], error) {
	endpoint := "/v5/asset/withdraw/query-record"

	params := S2M(req)
	var resp responses.BaseResponse[responses.GetWithdrawResponse]
	err := o.client.Do(ctx, http.MethodGet, endpoint, true, &resp, params)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
