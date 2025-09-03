//nolint:tagliatelle
package responses

import htxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/models"

type (
	WithdrawalVirtual struct {
		Basic
		WithdrawalTransferID int64 `json:"data,omitempty"`
	}
	GetWithdrawalAddress struct {
		Basic
		Addresses []*htxmodels.WithdrawalAddress `json:"data,omitempty"`
	}
	GetDepositAddress struct {
		Basic
		Addresses []*htxmodels.DepositAddress `json:"data,omitempty"`
	}
	CancelWithdrawal struct {
		Basic
		WithdrawalTransferID int64 `json:"data,omitempty"`
	}
	GetWithdrawalByClientID struct {
		Basic
		WithdrawalTransferData *htxmodels.WithdrawalByClientID `json:"data,omitempty"`
	}
	GetWithdrawalDepositHistory struct {
		Basic
		DepositWithdrawalData []*htxmodels.WithdrawalDepositHistory `json:"data,omitempty"`
	}
)
