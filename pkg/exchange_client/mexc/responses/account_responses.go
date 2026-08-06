package responses

import "github.com/shopspring/decimal"

type AccountBalance struct {
	Asset  string          `json:"asset"`
	Free   decimal.Decimal `json:"free"`
	Locked decimal.Decimal `json:"locked"`
}

type AccountInfoResponse struct {
	CanTrade    bool             `json:"canTrade"`
	CanWithdraw bool             `json:"canWithdraw"`
	CanDeposit  bool             `json:"canDeposit"`
	AccountType string           `json:"accountType"`
	Balances    []AccountBalance `json:"balances"`
	Permissions []string         `json:"permissions"`
}
