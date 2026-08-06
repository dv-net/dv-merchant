package responses

import "github.com/shopspring/decimal"

type CoinNetwork struct {
	Coin                    string          `json:"coin"`
	DepositDesc             string          `json:"depositDesc"`
	DepositEnable           bool            `json:"depositEnable"`
	MinConfirm              int64           `json:"minConfirm"`
	DepositPreConfirms      int64           `json:"depositPreConfirms"`
	Name                    string          `json:"name"`
	Network                 string          `json:"network"`
	WithdrawEnable          bool            `json:"withdrawEnable"`
	WithdrawFee             decimal.Decimal `json:"withdrawFee"`
	WithdrawIntegerMultiple decimal.Decimal `json:"withdrawIntegerMultiple"`
	WithdrawMax             decimal.Decimal `json:"withdrawMax"`
	WithdrawMin             decimal.Decimal `json:"withdrawMin"`
	SameAddress             bool            `json:"sameAddress"`
	Contract                string          `json:"contract"`
	WithdrawTips            *string         `json:"withdrawTips"`
	DepositTips             *string         `json:"depositTips"`
	NetWork                 string          `json:"netWork"`
}

type CoinConfig struct {
	Coin        string        `json:"coin"`
	Name        string        `json:"name"`
	NetworkList []CoinNetwork `json:"networkList"`
}

type GetCoinsConfigResponse []CoinConfig
