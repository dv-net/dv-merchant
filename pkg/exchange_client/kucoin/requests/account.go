//nolint:tagliatelle
package requests

import kucoinmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/models"

type GetAPIKeyInfo struct{}

type GetAccountList struct {
	Currency string `json:"currency,omitempty"`
	Type     string `json:"type,omitempty"`
}

type GetDepositAddress struct {
	Currency string `json:"currency" url:"currency"`
	Chain    string `json:"chain,omitempty" url:"chain,omitempty"`
}

type CreateDepositAddress struct {
	Currency string                   `json:"currency" url:"currency"`
	Chain    string                   `json:"chain,omitempty" url:"chain,omitempty"`
	To       kucoinmodels.AccountType `json:"to,omitempty" url:"to,omitempty"`
}

type GetWithdrawalHistory struct {
	Currency    string                        `json:"currency" url:"currency"`
	Status      kucoinmodels.WithdrawalStatus `json:"status,omitempty" url:"status,omitempty"`
	CurrentPage int64                         `json:"currentPage,omitempty" url:"currentPage,omitempty"`
	EndAt       int64                         `json:"endAt,omitempty" url:"endAt,omitempty"`
	PageSize    int64                         `json:"pageSize,omitempty" url:"pageSize,omitempty"`
	StartAt     int64                         `json:"startAt,omitempty" url:"startAt,omitempty"`
}

type CreateWithdrawal struct {
	Currency      string                     `json:"currency"`
	ToAddress     string                     `json:"toAddress"`
	Amount        string                     `json:"amount"`
	WithdrawType  kucoinmodels.WithdrawType  `json:"withdrawType"`
	Chain         string                     `json:"chain"`
	FeeDeductType kucoinmodels.FeeDeductType `json:"feeDeductType"`
}

type FlexTransfer struct {
	ClientOID       string                           `json:"clientOid"`
	Type            kucoinmodels.TransferType        `json:"type"`
	Currency        string                           `json:"currency"`
	Amount          string                           `json:"amount"`
	FromAccountType kucoinmodels.TransferAccountType `json:"fromAccountType"`
	ToAccountType   kucoinmodels.TransferAccountType `json:"toAccountType"`
}
