//nolint:tagliatelle
package responses

import (
	okxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/models"
)

type (
	ViewList struct {
		Basic
		SubAccounts []*okxmodels.SubAccount `json:"data,omitempty"`
	}
	APIKey struct {
		Basic
		APIKeys []*okxmodels.APIKey `json:"data,omitempty"`
	}
	GetSubAccountBalance struct {
		Basic
		Balances []*okxmodels.AccountBalance `json:"data,omitempty"`
	}
	HistoryTransfer struct {
		Basic
		HistoryTransfers []*okxmodels.HistoryTransfer `json:"data,omitempty"`
	}
	ManageTransfer struct {
		Basic
		Transfers []*okxmodels.FundingTransfer `json:"data,omitempty"`
	}
)
