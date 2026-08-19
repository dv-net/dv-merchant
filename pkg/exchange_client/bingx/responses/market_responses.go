package responses

import (
	"time"

	"github.com/shopspring/decimal"
)

type SymbolInfo struct {
	Symbol            string          `json:"symbol"`
	MinQty            decimal.Decimal `json:"minQty"`
	MaxQty            decimal.Decimal `json:"maxQty"`
	MinNotional       decimal.Decimal `json:"minNotional"`
	MaxNotional       decimal.Decimal `json:"maxNotional"`
	MaxMarketNotional decimal.Decimal `json:"maxMarketNotional"`
	Status            int             `json:"status"`
	TickSize          decimal.Decimal `json:"tickSize"`
	StepSize          decimal.Decimal `json:"stepSize"`
	APIStateSell      bool            `json:"apiStateSell"`
	APIStateBuy       bool            `json:"apiStateBuy"`
	TimeOnline        int64           `json:"timeOnline"`
	OffTime           int64           `json:"offTime"`
	MaintainTime      int64           `json:"maintainTime"`
	DisplayName       string          `json:"displayName"`
}

func (o SymbolInfo) IsDelisted() bool {
	return o.OffTime > 0 && o.OffTime <= time.Now().UnixMilli()
}

type GetSymbolsResponse struct {
	Symbols []SymbolInfo `json:"symbols"`
}

type TickerPrice struct {
	Symbol    string          `json:"symbol"`
	LastPrice decimal.Decimal `json:"lastPrice"`
	BidPrice  decimal.Decimal `json:"bidPrice"`
	AskPrice  decimal.Decimal `json:"askPrice"`
}

type GetTickerResponse []TickerPrice
