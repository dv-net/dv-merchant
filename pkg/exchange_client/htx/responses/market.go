//nolint:tagliatelle
package responses

import htxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/models"

// Market client responses
type (
	GetMarketTickers struct {
		Basic
		Tickers []*htxmodels.MarketTicker `json:"data,omitempty"`
		AdditionalData
	}
	GetMarketDetails struct {
		Channel string `json:"ch,omitempty"`
		Basic
		Details *htxmodels.MarketDetail `json:"tick,omitempty"`
		AdditionalData
	}
)
