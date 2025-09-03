//nolint:tagliatelle
package responses

import (
	htxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/models"
)

// Common client responses
type (
	GetSymbols struct {
		Basic
		Symbols []*htxmodels.Symbol `json:"data,omitempty"`
	}
	GetCurrencies struct {
		Basic
		Currencies []*htxmodels.Currency `json:"data,omitempty"`
		AdditionalData
	}
	GetCommonSymbols struct {
		Basic
		CommonSymbols []*htxmodels.CommonSymbol `json:"data,omitempty"`
	}
	GetMarketSymbols struct {
		Basic
		MarketSymbols []*htxmodels.MarketSymbol `json:"data,omitempty"`
	}
	GetCurrencyReference struct {
		Basic
		CurrencyReference []*htxmodels.CurrencyReference `json:"data,omitempty"`
	}
)
