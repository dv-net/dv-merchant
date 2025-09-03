//nolint:tagliatelle
package responses

import (
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/models"
)

type GetAPIKeyInfo struct {
	Basic
	Info *models.APIKeyInfo `json:"data"`
}

type GetAccountList struct {
	Basic
	Accounts []*models.Account `json:"data"`
}

type GetDepositAddress struct {
	Basic
	Addresses []*models.DepositAddress `json:"data"`
}

type CreateDepositAddress struct {
	Basic
	Address *models.DepositAddress `json:"data"`
}

type GetWithdrawalHistory struct {
	Basic
	History *models.WithdrawalHistory `json:"data"`
}

type CreateWithdrawal struct {
	Basic
	Data struct {
		WithdrawalID string `json:"withdrawalId"`
	} `json:"data"`
}

type FlexTransfer struct {
	Basic
	Data struct {
		OrderID string `json:"orderId"`
	} `json:"data"`
}
