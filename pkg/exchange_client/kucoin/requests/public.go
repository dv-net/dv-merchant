package requests

import (
	kucoinmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/models"
)

type GetAllSymbols struct {
	Market kucoinmodels.MarketType `json:"market,omitempty"`
}

type GetSymbol struct {
	Symbol string `json:"-"`
}

type GetTicker struct {
	Symbol string `json:"symbol" url:"symbol"`
}
