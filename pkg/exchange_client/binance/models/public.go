package models

import "github.com/shopspring/decimal"

type SymbolStatus string

func (o SymbolStatus) String() string { return string(o) }

const (
	SymbolStatusTrading SymbolStatus = "TRADING"
	SymbolStatusHalt    SymbolStatus = "HALT"
	SymbolStatusBreak   SymbolStatus = "BREAK"
)

type SymbolInfo struct {
	Symbol                          string        `json:"symbol"`
	Status                          SymbolStatus  `json:"status"`
	BaseAsset                       string        `json:"baseAsset"`
	BaseAssetPrecision              int           `json:"baseAssetPrecision"`
	QuoteAsset                      string        `json:"quoteAsset"`
	QuotePrecision                  int           `json:"quotePrecision"`
	QuoteAssetPrecision             int           `json:"quoteAssetPrecision"`
	OrderTypes                      []string      `json:"orderTypes"`
	IcebergAllowed                  bool          `json:"icebergAllowed"`
	OcoAllowed                      bool          `json:"ocoAllowed"`
	QuoteOrderQtyMarketAllowed      bool          `json:"quoteOrderQtyMarketAllowed"`
	AllowTrailingStop               bool          `json:"allowTrailingStop"`
	CancelReplaceAllowed            bool          `json:"cancelReplaceAllowed"`
	IsSpotTradingAllowed            bool          `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed          bool          `json:"isMarginTradingAllowed"`
	Filters                         []interface{} `json:"filters"`
	Permissions                     []interface{} `json:"permissions"`
	PermissionSets                  [][]string    `json:"permissionSets"`
	DefaultSelfTradePreventionMode  string        `json:"defaultSelfTradePreventionMode"`
	AllowedSelfTradePreventionModes []string      `json:"allowedSelfTradePreventionModes"`
}

type WalletType int

const (
	WalletTypeSpot WalletType = iota
	WalletTypeFunding
)

type MarketFilters struct {
	LotSizeFilter  *LotSizeFilter  `json:"lot_size_filter"`
	NotionalFilter *NotionalFilter `json:"notional_filter"`
}

type BaseFilter struct {
	FilterType string `json:"filterType"`
}

type LotSizeFilter struct {
	BaseFilter
	MinQty   decimal.Decimal `json:"minQty"` // Minimum order size in base asset
	MaxQty   decimal.Decimal `json:"maxQty"` // Maximum order size in base asset
	StepSize decimal.Decimal `json:"stepSize"`
}

type NotionalFilter struct {
	BaseFilter
	MinNotional      decimal.Decimal `json:"minNotional"` // Minimum order value in quote asset
	ApplyMinToMarket bool            `json:"applyMinToMarket"`
	MaxNotional      decimal.Decimal `json:"maxNotional"` // Maximum order value in quote asset
	ApplyMaxToMarket bool            `json:"applyMaxToMarket"`
	AvgPriceMins     int             `json:"avgPriceMins"`
}

type PriceFilter struct {
	FilterType string `json:"filterType"`
	MinPrice   string `json:"minPrice"`
	MaxPrice   string `json:"maxPrice"`
	TickSize   string `json:"tickSize"`
}

type PercentPriceFilter struct {
	FilterType     string `json:"filterType"`
	MultiplierUp   string `json:"multiplierUp"`
	MultiplierDown string `json:"multiplierDown"`
	AvgPriceMins   int    `json:"avgPriceMins"`
}

type PercentPriceBySideFilter struct {
	FilterType        string `json:"filterType"`
	BidMultiplierUp   string `json:"bidMultiplierUp"`
	BidMultiplierDown string `json:"bidMultiplierDown"`
	AskMultiplierUp   string `json:"askMultiplierUp"`
	AskMultiplierDown string `json:"askMultiplierDown"`
	AvgPriceMins      int    `json:"avgPriceMins"`
}

type MinNotionalFilter struct {
	FilterType    string `json:"filterType"`
	MinNotional   string `json:"minNotional"`
	ApplyToMarket bool   `json:"applyToMarket"`
	AvgPriceMins  int    `json:"avgPriceMins"`
}

type IcebergPartsFilter struct {
	FilterType string `json:"filterType"`
	Limit      int    `json:"limit"`
}

type MarketLotSizeFilter struct {
	FilterType string `json:"filterType"`
	MinQty     string `json:"minQty"`
	MaxQty     string `json:"maxQty"`
	StepSize   string `json:"stepSize"`
}

type MaxNumOrdersFilter struct {
	FilterType   string `json:"filterType"`
	MaxNumOrders int    `json:"maxNumOrders"`
}

type MaxNumAlgoOrdersFilter struct {
	FilterType       string `json:"filterType"`
	MaxNumAlgoOrders int    `json:"maxNumAlgoOrders"`
}

type MaxNumIcebergOrdersFilter struct {
	FilterType          string `json:"filterType"`
	MaxNumIcebergOrders int    `json:"maxNumIcebergOrders"`
}

type MaxPositionFilter struct {
	FilterType  string `json:"filterType"`
	MaxPosition string `json:"maxPosition"`
}

type TrailingDeltaFilter struct {
	FilterType            string `json:"filterType"`
	MinTrailingAboveDelta int    `json:"minTrailingAboveDelta"`
	MaxTrailingAboveDelta int    `json:"maxTrailingAboveDelta"`
	MinTrailingBelowDelta int    `json:"minTrailingBelowDelta"`
	MaxTrailingBelowDelta int    `json:"maxTrailingBelowDelta"`
}
