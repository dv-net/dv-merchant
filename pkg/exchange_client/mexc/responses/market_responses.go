package responses

import "github.com/shopspring/decimal"

type SymbolInfo struct {
	Symbol                     string   `json:"symbol"`
	Status                     string   `json:"status"`
	BaseAsset                  string   `json:"baseAsset"`
	BaseAssetPrecision         int      `json:"baseAssetPrecision"`
	QuoteAsset                 string   `json:"quoteAsset"`
	QuotePrecision             int      `json:"quotePrecision"`
	QuoteAssetPrecision        int      `json:"quoteAssetPrecision"`
	OrderTypes                 []string `json:"orderTypes"`
	IsSpotTradingAllowed       bool     `json:"isSpotTradingAllowed"`
	TradeSideType              int      `json:"tradeSideType"`
	BaseSizePrecision          string   `json:"baseSizePrecision"`
	QuoteAmountPrecision       string   `json:"quoteAmountPrecision"`
	QuoteAmountPrecisionMarket string   `json:"quoteAmountPrecisionMarket"`
	MaxQuoteAmount             string   `json:"maxQuoteAmount"`
	MaxQuoteAmountMarket       string   `json:"maxQuoteAmountMarket"`
}

type GetExchangeInfoResponse struct {
	Symbols []SymbolInfo `json:"symbols"`
}

type GetTickerPriceResponse struct {
	Symbol string          `json:"symbol"`
	Price  decimal.Decimal `json:"price"`
}
