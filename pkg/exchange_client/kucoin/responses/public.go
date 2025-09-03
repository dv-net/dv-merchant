//nolint:tagliatelle
package responses

import (
	kucoinmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/models"
)

type GetAllSymbols struct {
	Basic
	Symbols []*kucoinmodels.Symbol `json:"data"`
}

type GetSymbol struct {
	Basic
	Symbol *kucoinmodels.Symbol `json:"data"`
}

type GetTicker struct {
	Basic
	Ticker *kucoinmodels.Ticker `json:"data"`
}
