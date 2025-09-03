//nolint:tagliatelle
package requests

import (
	kucoinmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/models"
)

type GetOrderByOrderID struct {
	OrderID string `json:"-"`
	Symbol  string `json:"symbol" url:"symbol"`
}

type GetOrderByClientOID struct {
	ClientOrderOID string `json:"-"`
	Symbol         string `json:"symbol" url:"symbol"`
}

type CreateOrder struct {
	ClientOID string                            `json:"clientOid,omitempty"`
	Symbol    string                            `json:"symbol"`
	Type      kucoinmodels.OrderType            `json:"type"`
	Side      kucoinmodels.OrderSide            `json:"side"`
	Stp       kucoinmodels.SelfTradePreventType `json:"stp,omitempty"`
	Funds     string                            `json:"funds,omitempty"`
	Size      string                            `json:"size,omitempty"`
}
