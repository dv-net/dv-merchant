//nolint:tagliatelle
package responses

import (
	kucoinmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/models"
)

type GetOrderByClientOID struct {
	Basic
	Order *kucoinmodels.Order `json:"data"`
}

type GetOrderByOrderID struct {
	Basic
	Order *kucoinmodels.Order `json:"data"`
}

type CreateOrder struct {
	Basic
	Data struct {
		ClientOID string `json:"clientOid"`
		OrderID   string `json:"orderId"`
	}
}
