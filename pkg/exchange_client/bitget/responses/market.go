package responses

import "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/models"

type (
	TickerInformationResponse struct {
		CommonResponse
		Data []*models.TickerInformation `json:"data,omitempty"`
	}
	CoinInformationResponse struct {
		CommonResponse
		Data []*models.CoinInformation `json:"data,omitempty"`
	}
	SymbolInformationResponse struct {
		CommonResponse
		Data []*models.SymbolInformation `json:"data,omitempty"`
	}
)
