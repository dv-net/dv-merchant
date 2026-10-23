//nolint:tagliatelle
package responses

import (
	"fmt"

	"github.com/shopspring/decimal"
)

const (
	AssetStatusEnabled                    = "enabled"
	AssetStatusDepositOnly                = "deposit_only"
	AssetStatusWithdrawalOnly             = "withdrawal_only"
	AssetStatusFundingTemporarilyDisabled = "funding_temporarily_disabled"
)

const (
	PairStatusOnline     = "online"
	PairStatusCancelOnly = "cancel_only"
	PairStatusPostOnly   = "post_only"
	PairStatusLimitOnly  = "limit_only"
	PairStatusReduceOnly = "reduce_only"
)

type Asset struct {
	AClass          string `json:"aclass"`
	AltName         string `json:"altname"` // "XBT" for "XXBT", "USD" for "ZUSD"
	Decimals        int    `json:"decimals"`
	DisplayDecimals int    `json:"display_decimals"`
	Status          string `json:"status"`
}

func (a Asset) CanDeposit() bool {
	return a.Status == AssetStatusEnabled || a.Status == AssetStatusDepositOnly
}

func (a Asset) CanWithdraw() bool {
	return a.Status == AssetStatusEnabled || a.Status == AssetStatusWithdrawalOnly
}

// AssetsResponse is keyed by Kraken asset name ("XXBT", "ZUSD", "USDT", ...).
type AssetsResponse map[string]Asset

type AssetPair struct {
	AltName      string          `json:"altname"` // "XBTUSDT"
	WsName       string          `json:"wsname"`  // "XBT/USDT"
	Base         string          `json:"base"`    // "XXBT"
	Quote        string          `json:"quote"`   // "USDT"
	PairDecimals int             `json:"pair_decimals"`
	CostDecimals int             `json:"cost_decimals"`
	LotDecimals  int             `json:"lot_decimals"`
	OrderMin     decimal.Decimal `json:"ordermin"` // min volume in base currency
	CostMin      decimal.Decimal `json:"costmin"`  // min cost in quote currency
	TickSize     decimal.Decimal `json:"tick_size"`
	Status       string          `json:"status"`
}

func (p AssetPair) IsTradable() bool {
	return p.Status == PairStatusOnline
}

// AssetPairsResponse is keyed by Kraken pair name ("XXBTZUSD", "XBTUSDT", ...).
type AssetPairsResponse map[string]AssetPair

// Ticker arrays follow Kraken format:
// a/b: [price, whole lot volume, lot volume], c: [price, lot volume],
// v/p/l/h: [today, last 24 hours], t: [today, last 24 hours].
type Ticker struct {
	Ask       []decimal.Decimal `json:"a"`
	Bid       []decimal.Decimal `json:"b"`
	LastTrade []decimal.Decimal `json:"c"`
	Volume    []decimal.Decimal `json:"v"`
	VWAP      []decimal.Decimal `json:"p"`
	Trades    []int64           `json:"t"`
	Low       []decimal.Decimal `json:"l"`
	High      []decimal.Decimal `json:"h"`
	Open      decimal.Decimal   `json:"o"`
}

func (t Ticker) AskPrice() (decimal.Decimal, error)  { return first(t.Ask, "ask") }
func (t Ticker) BidPrice() (decimal.Decimal, error)  { return first(t.Bid, "bid") }
func (t Ticker) LastPrice() (decimal.Decimal, error) { return first(t.LastTrade, "last trade") }

func (t Ticker) Volume24h() decimal.Decimal {
	if len(t.Volume) < 2 {
		return decimal.Zero
	}
	return t.Volume[1]
}

// TickerResponse is keyed by Kraken pair name ("XXBTZUSD", ...),
// even if the request used the alternative name ("XBTUSD").
type TickerResponse map[string]Ticker

func first(values []decimal.Decimal, name string) (decimal.Decimal, error) {
	if len(values) == 0 {
		return decimal.Zero, fmt.Errorf("kraken ticker: %s price is missing", name)
	}
	return values[0], nil
}
