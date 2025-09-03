//nolint:tagliatelle
package models

type MarketTicker struct {
	Amount  float64 `json:"amount"`
	Count   int     `json:"count"`
	Open    float64 `json:"open"`
	Close   float64 `json:"close"`
	Low     float64 `json:"low"`
	High    float64 `json:"high"`
	Volume  float64 `json:"vol"`
	Symbol  string  `json:"symbol"`
	Bid     float64 `json:"bid"`
	BidSize float64 `json:"bidSize"`
	Ask     float64 `json:"ask"`
	AskSize float64 `json:"askSize"`
}

type MarketDetail struct {
	ID      int     `json:"id"`
	Amount  float64 `json:"amount"`
	Count   int     `json:"count"`
	Open    float64 `json:"open"`
	Close   float64 `json:"close"`
	Low     float64 `json:"low"`
	High    float64 `json:"high"`
	Volume  float64 `json:"vol"`
	Version int     `json:"version"`
}
