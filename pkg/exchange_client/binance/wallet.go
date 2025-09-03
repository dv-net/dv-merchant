package binance

import (
	"context"
	"net/http"
	"net/url"

	binancerequests "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/requests"
	binanceresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/responses"
)

type IWalletClient interface {
	GetDefaultDepositAddress(ctx context.Context, req *binancerequests.GetDefaultDepositAddressRequest) (*binanceresponses.GetDefaultDepositAddressResponse, error)
	Withdrawal(ctx context.Context, request *binancerequests.WithdrawalRequest) (*binanceresponses.WithdrawalResponse, error)
	GetAllCoinInformation(ctx context.Context) (*binanceresponses.GetAllCoinInformationResponse, error)
	GetFundingAssets(ctx context.Context, request *binancerequests.GetFundingAssetsRequest) (*binanceresponses.GetFundingAssetsResponse, error)
	GetSpotAssets(ctx context.Context, request *binancerequests.GetSpotAssetsRequest) (*binanceresponses.GetSpotAssetsResponse, error)
	GetDepositAddresses(ctx context.Context, request *binancerequests.GetDepositAddressesRequest) (*binanceresponses.GetDepositAddressesResponse, error)
	GetUserBalances(ctx context.Context, request *binancerequests.GetUserBalancesRequest) (*binanceresponses.GetUserBalancesResponse, error)
	GetWithdrawalAddresses(ctx context.Context) (*binanceresponses.GetWithdrawalAddressesResponse, error)
	GetWithdrawalHistory(ctx context.Context, request *binancerequests.GetWithdrawalHistoryRequest) (*binanceresponses.GetWithdrawalHistoryResponse, error)
	UniversalTransfer(ctx context.Context, request *binancerequests.UniversalTransferRequest) (*binanceresponses.UniversalTransferResponse, error)
}

func NewWallet(opt *ClientOptions) (IWalletClient, error) {
	client, err := NewClient(opt)
	if err != nil {
		return nil, err
	}
	wallet := &WalletClient{
		client: client,
	}

	return wallet, nil
}

type WalletClient struct {
	client *Client
}

func (o *WalletClient) GetSpotAssets(ctx context.Context, request *binancerequests.GetSpotAssetsRequest) (*binanceresponses.GetSpotAssetsResponse, error) {
	response := &binanceresponses.GetSpotAssetsResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/sapi/v3/asset/getUserAsset")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelSigned, &response.Data); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetWithdrawalHistory(ctx context.Context, request *binancerequests.GetWithdrawalHistoryRequest) (*binanceresponses.GetWithdrawalHistoryResponse, error) {
	response := &binanceresponses.GetWithdrawalHistoryResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/sapi/v1/capital/withdraw/history")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelSigned, &response.Data); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetDefaultDepositAddress(ctx context.Context, request *binancerequests.GetDefaultDepositAddressRequest) (*binanceresponses.GetDefaultDepositAddressResponse, error) {
	response := &binanceresponses.GetDefaultDepositAddressResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/sapi/v1/capital/deposit/address")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelSigned, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) Withdrawal(ctx context.Context, request *binancerequests.WithdrawalRequest) (*binanceresponses.WithdrawalResponse, error) {
	response := &binanceresponses.WithdrawalResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/sapi/v1/capital/withdraw/apply")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelSigned, response); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetAllCoinInformation(ctx context.Context) (*binanceresponses.GetAllCoinInformationResponse, error) {
	response := &binanceresponses.GetAllCoinInformationResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/sapi/v1/capital/config/getall")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelSigned, &response.Data); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetFundingAssets(ctx context.Context, request *binancerequests.GetFundingAssetsRequest) (*binanceresponses.GetFundingAssetsResponse, error) {
	response := &binanceresponses.GetFundingAssetsResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/sapi/v1/asset/get-funding-asset")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelSigned, &response.Data); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetDepositAddresses(ctx context.Context, request *binancerequests.GetDepositAddressesRequest) (*binanceresponses.GetDepositAddressesResponse, error) {
	response := &binanceresponses.GetDepositAddressesResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/sapi/v1/capital/deposit/address/list")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelSigned, &response.Data); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetUserBalances(ctx context.Context, request *binancerequests.GetUserBalancesRequest) (*binanceresponses.GetUserBalancesResponse, error) {
	response := &binanceresponses.GetUserBalancesResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/sapi/v1/asset/wallet/balance")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelSigned, &response.Data); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) GetWithdrawalAddresses(ctx context.Context) (*binanceresponses.GetWithdrawalAddressesResponse, error) {
	response := &binanceresponses.GetWithdrawalAddressesResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/sapi/v1/capital/withdraw/address/list")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelSigned, &response.Data); err != nil {
		return nil, err
	}
	return response, nil
}

func (o *WalletClient) UniversalTransfer(ctx context.Context, request *binancerequests.UniversalTransferRequest) (*binanceresponses.UniversalTransferResponse, error) {
	response := &binanceresponses.UniversalTransferResponse{}

	path, err := url.JoinPath(o.client.baseURL.String(), "/sapi/v1/asset/transfer")
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, http.NoBody)
	if err != nil {
		return nil, err
	}
	req, err = o.client.assembleRequest(request, req)
	if err != nil {
		return nil, err
	}
	if err := o.client.Do(ctx, req, SecurityLevelSigned, response); err != nil {
		return nil, err
	}
	return response, nil
}
