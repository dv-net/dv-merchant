//nolint:tagliatelle
package models

import (
	"github.com/shopspring/decimal"
)

type TransactionLog struct {
	TransSubType    string          `json:"transSubType,omitempty"`
	ID              string          `json:"id"`
	Symbol          string          `json:"symbol"`
	Side            Side            `json:"side"`
	Funding         string          `json:"funding"`
	OrderLinkID     string          `json:"orderLinkId"`
	OrderID         string          `json:"orderId"`
	Fee             string          `json:"fee"`
	Change          string          `json:"change"`
	CashFlow        string          `json:"cashFlow"`
	TransactionTime string          `json:"transactionTime"`
	Type            string          `json:"type"`
	FeeRate         decimal.Decimal `json:"feeRate"`
	BonusChange     string          `json:"bonusChange,omitempty"`
	Size            decimal.Decimal `json:"size"`
	Qty             decimal.Decimal `json:"qty"`
	CashBalance     string          `json:"cashBalance"`
	Currency        string          `json:"currency"`
	Category        string          `json:"category"`
	TradePrice      string          `json:"tradePrice"`
	TradeID         string          `json:"tradeId"`
	ExtraFees       string          `json:"extraFees,omitempty"`
}

type TradingBalance struct {
	AccountType string            `json:"accountType"`
	Coin        []CoinBalanceInfo `json:"coin"`
}

type CoinBalanceInfo struct {
	Coin          string `json:"coin"`
	WalletBalance string `json:"walletBalance,omitempty"`
}

type AccountBalance struct {
	Coin            string `json:"coin"`
	WalletBalance   string `json:"walletBalance,omitempty"`
	TransferBalance string `json:"transferBalance,omitempty"`
}

type DepositChain struct {
	ChainType      string `json:"chainType"`
	AddressDeposit string `json:"addressDeposit"`
	TagDeposit     string `json:"tagDeposit,omitempty"`
	Chain          string `json:"chain"`
}

type Withdraw struct {
	Coin         string           `json:"coin"`
	Chain        string           `json:"chain"`
	Amount       decimal.Decimal  `json:"amount"`
	TxID         string           `json:"txID,omitempty"`
	Status       WithdrawalStatus `json:"status"`
	ToAddress    string           `json:"toAddress"`
	Tag          string           `json:"tag,omitempty"`
	WithdrawFee  string           `json:"withdrawFee"`
	CreateTime   string           `json:"createTime"`
	UpdateTime   string           `json:"updateTime"`
	WithdrawID   string           `json:"withdrawId"`
	WithdrawType WithdrawType     `json:"withdrawType"`
}
