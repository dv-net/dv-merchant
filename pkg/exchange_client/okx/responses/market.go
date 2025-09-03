//nolint:tagliatelle
package responses

import (
	okxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/models"
)

type (
	Ticker struct {
		Basic
		Tickers []*okxmodels.Ticker `json:"data,omitempty"`
	}
	IndexTicker struct {
		Basic
		IndexTickers []*okxmodels.IndexTicker `json:"data,omitempty"`
	}
	OrderBook struct {
		Basic
		OrderBooks []*okxmodels.OrderBook `json:"data,omitempty"`
	}
	Candle struct {
		Basic
		Candles []*okxmodels.Candle `json:"data,omitempty"`
	}
	IndexCandle struct {
		Basic
		Candles []*okxmodels.IndexCandle `json:"data,omitempty"`
	}
	CandleMarket struct {
		Basic
		Candles []*okxmodels.IndexCandle `json:"data,omitempty"`
	}
	Trade struct {
		Basic
		Trades []*okxmodels.Trade `json:"data,omitempty"`
	}
	TotalVolume24H struct {
		Basic
		TotalVolume24Hs []*okxmodels.TotalVolume24H `json:"data,omitempty"`
	}
	IndexComponent struct {
		Basic
		IndexComponents *okxmodels.IndexComponent `json:"data,omitempty"`
	}
)
