package requests

import "github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/models"

type GetInstrumentsRequest struct {
	Category string                  `json:"category" url:"category"` // Category of the instrument, e.g., "spot", "linear", "inverse"
	Symbol   string                  `json:"symbol,omitempty" url:"symbol,omitempty"`
	Status   models.InstrumentStatus `json:"status,omitempty" url:"status,omitempty"`
}

type GetTickersRequest struct {
	Category string `json:"category" url:"category"`                 // Category of the instrument, e.g., "spot", "linear", "inverse"
	Symbol   string `json:"symbol,omitempty" url:"symbol,omitempty"` // Specific symbol to get ticker for
}

type GetCoinInfoRequest struct {
	Coin string `json:"coin,omitempty" url:"coin,omitempty"`
}
