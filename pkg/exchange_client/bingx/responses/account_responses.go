package responses

import "github.com/shopspring/decimal"

type AccountBalance struct {
	Asset  string          `json:"asset"`
	Free   decimal.Decimal `json:"free"`
	Locked decimal.Decimal `json:"locked"`
}

type GetBalanceResponse struct {
	Balances []AccountBalance `json:"balances"`
}

type FundAsset struct {
	Asset  string          `json:"asset"`
	Free   decimal.Decimal `json:"free"`
	Locked decimal.Decimal `json:"locked"`
}

type GetFundBalanceResponse struct {
	Assets []FundAsset `json:"assets"`
}

type TransferResponse struct {
	TranID int64 `json:"tranId"`
}
