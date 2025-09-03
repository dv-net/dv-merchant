//nolint:tagliatelle
package responses

import (
	okxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/models"
)

type (
	GetAccountBalance struct {
		Basic
		Balances []*okxmodels.AccountBalance `json:"data,omitempty"`
	}
	GetPositions struct {
		Basic
		Positions []*okxmodels.Position `json:"data"`
	}
	GetAccountAndPositionRisk struct {
		Basic
		PositionAndAccountRisks []*okxmodels.PositionAndAccountRisk `json:"data"`
	}
	GetBills struct {
		Basic
		Bills []*okxmodels.Bill `json:"data"`
	}
	GetConfig struct {
		Basic
		Configs []*okxmodels.Config `json:"data"`
	}
	SetPositionMode struct {
		Basic
		PositionModes []*okxmodels.PositionMode `json:"data"`
	}
	Leverage struct {
		Basic
		Leverages []*okxmodels.Leverage `json:"data"`
	}
	GetMaxBuySellAmount struct {
		Basic
		MaxBuySellAmounts []*okxmodels.MaxBuySellAmount `json:"data"`
	}
	GetMaxAvailableTradeAmount struct {
		Basic
		MaxAvailableTradeAmounts []*okxmodels.MaxAvailableTradeAmount `json:"data"`
	}
	IncreaseDecreaseMargin struct {
		Basic
		MarginBalanceAmounts []*okxmodels.MarginBalanceAmount `json:"data"`
	}
	GetMaxLoan struct {
		Basic
		Loans []*okxmodels.Loan `json:"data"`
	}
	GetFeeRates struct {
		Basic
		Fees []*okxmodels.Fee `json:"data"`
	}
	GetInterestAccrued struct {
		Basic
		InterestAccrues []*okxmodels.InterestAccrued `json:"data"`
	}
	GetInterestRates struct {
		Basic
		Interests []*okxmodels.InterestRate `json:"data"`
	}
	SetGreeks struct {
		Basic
		Greeks []*okxmodels.Greek `json:"data"`
	}
	GetMaxWithdrawals struct {
		Basic
		MaxWithdrawals []*okxmodels.MaxWithdrawal `json:"data"`
	}
)
