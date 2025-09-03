package responses

import "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/models"

type (
	AccountAssetsResponse struct {
		CommonResponse
		Data []*models.AccountAsset `json:"data,omitempty"`
	}
	DepositAddressResponse struct {
		CommonResponse
		Data *models.DepositAddress `json:"data,omitempty"`
	}
	DepositRecordsResponse struct {
		CommonResponse
		Data []*models.DepositRecord `json:"data,omitempty"`
	}
	WithdrawalRecordsResponse struct {
		CommonResponse
		Data []*models.WithdrawalRecord `json:"data,omitempty"`
	}
	WalletWithdrawalResponse struct {
		CommonResponse
		Data *models.Withdrawal `json:"data,omitempty"`
	}
)
