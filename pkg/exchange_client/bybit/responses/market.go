package responses

import "github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/models"

type GetInstrumentsResponse struct {
	Category string              `json:"category"`
	List     []models.Instrument `json:"list"`
}

type GetTickersResponse struct {
	Category string          `json:"category"`
	List     []models.Ticker `json:"list"`
}

type GetCoinInfoResponse struct {
	Rows []models.CoinInfo `json:"rows"`
}
