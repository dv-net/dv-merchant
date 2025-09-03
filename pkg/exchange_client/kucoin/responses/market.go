//nolint:tagliatelle
package responses

import (
	kucoinmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/models"
)

type GetCurrencyList struct {
	Basic
	Currencies []*kucoinmodels.Currency `json:"data"`
}

type GetCurrency struct {
	Basic
	Currency *kucoinmodels.Currency `json:"data"`
}
