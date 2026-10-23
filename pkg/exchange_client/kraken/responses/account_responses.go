package responses

import "github.com/shopspring/decimal"

// BalanceResponse maps Kraken asset names ("XXBT", "ZUSD", "USDT", "USDT.F", ...) to total balance.
type BalanceResponse map[string]decimal.Decimal

type BalanceEx struct {
	Balance    decimal.Decimal `json:"balance"`
	HoldTrade  decimal.Decimal `json:"hold_trade"`
	Credit     decimal.Decimal `json:"credit"`
	CreditUsed decimal.Decimal `json:"credit_used"`
}

// Available is the amount that can be traded or withdrawn right now,
// as defined by Kraken: balance + credit - credit_used - hold_trade.
func (b BalanceEx) Available() decimal.Decimal {
	return b.Balance.Add(b.Credit).Sub(b.CreditUsed).Sub(b.HoldTrade)
}

// BalanceExResponse maps Kraken asset names to extended balance info.
type BalanceExResponse map[string]BalanceEx
