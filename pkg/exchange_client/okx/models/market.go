//nolint:tagliatelle
package models

type (
	Ticker struct {
		InstID    string `json:"instId"`
		Last      string `json:"last"`
		LastSz    string `json:"lastSz"`
		AskPx     string `json:"askPx"`
		AskSz     string `json:"askSz"`
		BidPx     string `json:"bidPx"`
		BidSz     string `json:"bidSz"`
		Open24h   string `json:"open24h"`
		High24h   string `json:"high24h"`
		Low24h    string `json:"low24h"`
		VolCcy24h string `json:"volCcy24h"`
		Vol24h    string `json:"vol24h"`
		SodUtc0   string `json:"sodUtc0"`
		SodUtc8   string `json:"sodUtc8"`
		InstType  string `json:"instType"`
		TS        string `json:"ts"`
	}
	IndexTicker struct {
		InstID    string `json:"instId"`
		IdxPx     string `json:"idxPx"`
		HighDay   string `json:"high24h"`
		SodUTC0   string `json:"sodUtc0"`
		OpenDay   string `json:"open24h"`
		LowDay    string `json:"low24h"`
		SodUTC8   string `json:"sodUtc8"`
		Timestamp string `json:"ts"`
	}
	OrderBook struct {
		Asks []*OrderBookEntity `json:"asks"`
		Bids []*OrderBookEntity `json:"bids"`
		TS   int64              `json:"ts"`
	}
	OrderBookWs struct {
		Asks     []*OrderBookEntity `json:"asks"`
		Bids     []*OrderBookEntity `json:"bids"`
		Checksum int                `json:"checksum"`
		TS       int64              `json:"ts"`
	}
	OrderBookEntity struct {
		DepthPrice      float64
		Size            float64
		LiquidatedOrder int
		OrderNumbers    int
	}
	Candle struct {
		O      float64
		H      float64
		L      float64
		C      float64
		Vol    float64
		VolCcy float64
		TS     int64
	}
	IndexCandle struct {
		O  float64
		H  float64
		L  float64
		C  float64
		TS int64
	}
	Trade struct {
		InstID  string `json:"instId"`
		TradeID string `json:"tradeId"`
		Px      string `json:"px"`
		Sz      string `json:"sz"`
		Side    string `json:"side"`
		TS      int64  `json:"ts"`
	}
	TotalVolume24H struct {
		VolUsd string `json:"volUsd"`
		VolCny string `json:"volCny"`
		TS     int64  `json:"ts"`
	}
	IndexComponent struct {
		Index      string       `json:"index"`
		Last       string       `json:"last"`
		Components []*Component `json:"components"`
		TS         int64        `json:"ts"`
	}
	Component struct {
		Exch   string `json:"exch"`
		Symbol string `json:"symbol"`
		SymPx  string `json:"symPx"`
		Wgt    string `json:"wgt"`
		CnvPx  string `json:"cnvPx"`
	}
)
