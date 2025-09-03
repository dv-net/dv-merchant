//nolint:tagliatelle
package models

import "github.com/shopspring/decimal"

type Symbol struct {
	Symbol         string          `json:"symbol"`
	Name           string          `json:"name"`
	BaseCurrency   string          `json:"baseCurrency"`
	QuoteCurrency  string          `json:"quoteCurrency"`
	FeeCurrency    string          `json:"feeCurrency"`
	Market         string          `json:"market"`
	BaseMinSize    decimal.Decimal `json:"baseMinSize"`
	QuoteMinSize   decimal.Decimal `json:"quoteMinSize"`
	BaseMaxSize    decimal.Decimal `json:"baseMaxSize"`
	QuoteMaxSize   decimal.Decimal `json:"quoteMaxSize"`
	BaseIncrement  decimal.Decimal `json:"baseIncrement"`
	QuoteIncrement decimal.Decimal `json:"quoteIncrement"`
	PriceIncrement decimal.Decimal `json:"priceIncrement"`
	PriceLimitRate decimal.Decimal `json:"priceLimitRate"`
	MinFunds       decimal.Decimal `json:"minFunds"`
	EnableTrading  bool            `json:"enableTrading"`
}

type Ticker struct {
	BestAsk     decimal.Decimal `json:"bestAsk"`
	BestAskSize decimal.Decimal `json:"bestAskSize"`
	BestBid     decimal.Decimal `json:"bestBid"`
	BestBidSize decimal.Decimal `json:"bestBidSize"`
	Price       decimal.Decimal `json:"price"`
	Sequence    string          `json:"sequence"`
	Size        decimal.Decimal `json:"size"`
	Time        int64           `json:"time"`
}
