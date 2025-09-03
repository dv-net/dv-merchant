package requests

import (
	binancemodels "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/models"
)

type GetExchangeInfoRequest struct {
	Symbol             string                       `json:"symbol,omitempty" url:"symbol,omitempty"`
	Symbols            []binancemodels.SymbolInfo   `json:"symbols,omitempty" url:"symbols,omitempty"`
	Permissions        []string                     `json:"permissions,omitempty" url:"permissions,omitempty"`
	ShowPermissionSets bool                         `json:"showPermissionSets,omitempty" url:"showPermissionSets,omitempty"`
	SymbolStatus       []binancemodels.SymbolStatus `json:"symbolStatus,omitempty" url:"symbolStatus,omitempty"`
}

type GetSymbolPriceTickerRequest struct {
	Symbol string `json:"symbol,omitempty" url:"symbol"`
}

type GetSymbolsPriceTickerRequest struct {
	Symbols []string `json:"symbols,omitempty" url:"symbols,brackets,comma,omitempty"` //fixme: brackets
}
