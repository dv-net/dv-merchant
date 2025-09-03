package responses

import binancemodels "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/models"

type GetPingResponse struct{}

type GetServerTimeResponse struct {
	ServerTime int64 `json:"serverTime"`
}

type GetExchangeInfoResponse struct {
	Timezone        string                     `json:"timezone"`
	ServerTime      int64                      `json:"serverTime"`
	RateLimits      []struct{}                 `json:"rateLimits"`
	ExchangeFilters []interface{}              `json:"exchangeFilters"`
	Symbols         []binancemodels.SymbolInfo `json:"symbols"`
}

type GetSymbolPriceTickerResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type Symbols []GetSymbolPriceTickerResponse

type GetSymbolsPriceTickerResponse struct {
	Data Symbols
}
