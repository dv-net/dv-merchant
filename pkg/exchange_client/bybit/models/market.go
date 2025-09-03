//nolint:tagliatelle
package models

import "github.com/shopspring/decimal"

type Instrument struct {
	Symbol        string           `json:"symbol"`
	BaseCoin      string           `json:"baseCoin"`
	QuoteCoin     string           `json:"quoteCoin"`
	Status        InstrumentStatus `json:"status"`
	LotSizeFilter struct {
		BasePrecision  string          `json:"basePrecision"`
		QuotePrecision string          `json:"quotePrecision"`
		MinOrderQty    decimal.Decimal `json:"minOrderQty"`
		MaxOrderQty    decimal.Decimal `json:"maxOrderQty"`
		MinOrderAmt    decimal.Decimal `json:"minOrderAmt"`
		MaxOrderAmt    decimal.Decimal `json:"maxOrderAmt"`
	} `json:"lotSizeFilter"`
	PriceFilter struct {
		TickSize string `json:"tickSize"`
	}
}

type Ticker struct {
	Symbol    string `json:"symbol"`
	Bid1Price string `json:"bid1Price"`
	Bid1Size  string `json:"bid1Size"`
	Ask1Price string `json:"ask1Price"`
	Ask1Size  string `json:"ask1Size"`
	LastPrice string `json:"lastPrice"`
}

type ChainInfo struct {
	ChainType             string `json:"chainType"`
	Confirmation          string `json:"confirmation"`
	WithdrawFee           string `json:"withdrawFee"`
	DepositMin            string `json:"depositMin"`
	WithdrawMin           string `json:"withdrawMin"`
	Chain                 string `json:"chain"`
	ChainDeposit          string `json:"chainDeposit"`
	ChainWithdraw         string `json:"chainWithdraw"`
	MinAccuracy           string `json:"minAccuracy"`
	WithdrawPercentageFee string `json:"withdrawPercentageFee"`
	ContractAddress       string `json:"contractAddress,omitempty"`
	SafeConfirmNumber     string `json:"safeConfirmNumber"`
}

type CoinInfo struct {
	Name         string          `json:"name"`
	Coin         string          `json:"coin"`
	RemainAmount decimal.Decimal `json:"remainAmount"`
	Chains       []*ChainInfo    `json:"chains"`
}
