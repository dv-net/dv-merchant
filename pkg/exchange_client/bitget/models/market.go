//nolint:tagliatelle
package models

import "github.com/shopspring/decimal"

type (
	TickerInformation struct {
		Symbol       string          `json:"symbol"`
		High24h      decimal.Decimal `json:"high24h"`
		Open         decimal.Decimal `json:"open"`
		Low24h       decimal.Decimal `json:"low24h"`
		LastPr       decimal.Decimal `json:"lastPr"`
		QuoteVolume  decimal.Decimal `json:"quoteVolume"`
		BaseVolume   decimal.Decimal `json:"baseVolume"`
		UsdtVolume   decimal.Decimal `json:"usdtVolume"`
		BidPr        decimal.Decimal `json:"bidPr"`
		AskPr        decimal.Decimal `json:"askPr"`
		BidSz        decimal.Decimal `json:"bidSz"`
		AskSz        decimal.Decimal `json:"askSz"`
		OpenUtc      decimal.Decimal `json:"openUtc"`
		TS           int             `json:"ts,string"`
		ChangeUtc24h decimal.Decimal `json:"changeUtc24h"`
		Change24h    decimal.Decimal `json:"change24h"`
	}

	CoinInformation struct {
		CoinID   string                 `json:"coinId"`
		Coin     string                 `json:"coin"`
		Transfer bool                   `json:"transfer,string"`
		Chains   []CoinChainInformation `json:"chains"`
	}

	CoinChainInformation struct {
		Chain             string          `json:"chain"`
		NeedTag           bool            `json:"needTag,string"`
		Withdrawable      bool            `json:"withdrawable,string"`
		Rechargeable      bool            `json:"rechargeable,string"`
		WithdrawFee       decimal.Decimal `json:"withdrawFee"`
		ExtraWithdrawFee  decimal.Decimal `json:"extraWithdrawFee"`
		DepositConfirm    decimal.Decimal `json:"depositConfirm"`
		WithdrawConfirm   decimal.Decimal `json:"withdrawConfirm"`
		MinDepositAmount  decimal.Decimal `json:"minDepositAmount"`
		MinWithdrawAmount decimal.Decimal `json:"minWithdrawAmount"`
		BrowserURL        string          `json:"browserUrl"`
		ContractAddress   string          `json:"contractAddress"`
		WithdrawStep      decimal.Decimal `json:"withdrawStep"`
		WithdrawMinScale  decimal.Decimal `json:"withdrawMinScale"`
		Congestion        string          `json:"congestion"`
	}

	SymbolInformation struct {
		Symbol            string          `json:"symbol"`
		BaseCoin          string          `json:"baseCoin"`
		QuoteCoin         string          `json:"quoteCoin"`
		MinTradeAmount    decimal.Decimal `json:"minTradeAmount"`
		MaxTradeAmount    decimal.Decimal `json:"maxTradeAmount"`
		TakerFeeRate      decimal.Decimal `json:"takerFeeRate"`
		MakerFeeRate      decimal.Decimal `json:"makerFeeRate"`
		PricePrecision    int             `json:"pricePrecision,string"`
		QuantityPrecision int             `json:"quantityPrecision,string"`
		QuotePrecision    int             `json:"quotePrecision,string"`
		MinTradeUSDT      decimal.Decimal `json:"minTradeUSDT"`
		Status            SymbolStatus    `json:"status"`
	}
)

type SymbolStatus string

func (o SymbolStatus) String() string { return string(o) }

const (
	SymbolStatusOnline  SymbolStatus = "online"
	SymbolStatusOffline SymbolStatus = "offline"
	SymbolStatusHalt    SymbolStatus = "halt"
	SymbolStatusGrey    SymbolStatus = "gray"
)
