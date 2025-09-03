//nolint:tagliatelle
package models

import "github.com/shopspring/decimal"

type Order struct {
	ID          string          `json:"id"`
	ClientOID   string          `json:"clientOid,omitempty"`
	Symbol      string          `json:"symbol"`
	Active      bool            `json:"active"`
	Type        OrderType       `json:"type"`
	DealFunds   decimal.Decimal `json:"dealFunds"`
	DealSize    decimal.Decimal `json:"dealSize"`
	Fee         decimal.Decimal `json:"fee"`
	FeeCurrency string          `json:"feeCurrency"`
	Funds       decimal.Decimal `json:"funds"`
	InOrderBook bool            `json:"inOrderBook"`
	OpType      string          `json:"opType"`
	Price       decimal.Decimal `json:"price"`
	RemainFunds decimal.Decimal `json:"remainFunds"`
	RemainSize  decimal.Decimal `json:"remainSize"`
	Side        OrderSide       `json:"side"`
	Size        decimal.Decimal `json:"size"`
}
