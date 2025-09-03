//nolint:tagliatelle
package models

import (
	"github.com/shopspring/decimal"
)

type (
	OrderInformation struct {
		UserID           string          `json:"userId"`
		Symbol           string          `json:"symbol"`
		OrderID          string          `json:"orderId"`
		ClientOID        string          `json:"clientOid"`
		Price            decimal.Decimal `json:"price"`
		Size             decimal.Decimal `json:"size"`
		OrderType        OrderType       `json:"orderType"`
		Side             string          `json:"side"`
		Status           OrderStatus     `json:"status"`
		PriceAvg         decimal.Decimal `json:"priceAvg"`
		BaseVolume       decimal.Decimal `json:"baseVolume"`
		QuoteVolume      decimal.Decimal `json:"quoteVolume"`
		EnterPointSource string          `json:"enterPointSource"`
		FeeDetail        string          `json:"feeDetail"`
		OrderSource      string          `json:"orderSource"`
		CancelReason     string          `json:"cancelReason,omitempty"`
		CTime            int             `json:"cTime,string"`
		UTime            int             `json:"uTime,string"`
	}
	PlacedOrder struct {
		OrderID   string `json:"orderId,omitempty"`
		ClientOID string `json:"clientOid,omitempty"`
	}
)
