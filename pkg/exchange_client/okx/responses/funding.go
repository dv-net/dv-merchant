//nolint:tagliatelle
package responses

import (
	okxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/models"
)

type (
	GetCurrencies struct {
		Basic
		Currencies []*okxmodels.Currency `json:"data"`
	}
	GetFundingBalance struct {
		Basic
		Balances []*okxmodels.FundingBalance `json:"data"`
	}
	FundsTransfer struct {
		Basic
		Transfers []*okxmodels.FundingTransfer `json:"data"`
	}
	AssetBillsDetails struct {
		Basic
		Bills []*okxmodels.Bill `json:"data"`
	}
	GetDepositAddress struct {
		Basic
		DepositAddresses []*okxmodels.DepositAddress `json:"data"`
	}
	GetDepositHistory struct {
		Basic
		DepositHistories []*okxmodels.DepositHistory `json:"data"`
	}
	Withdrawal struct {
		Basic
		Withdrawals []*okxmodels.Withdrawal `json:"data"`
	}
	GetWithdrawalHistory struct {
		Basic
		WithdrawalHistories []*okxmodels.WithdrawalHistory `json:"data"`
	}
)
