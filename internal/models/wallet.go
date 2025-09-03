package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WalletWithUSDBalance struct {
	WalletAddressID uuid.UUID       `json:"wallet_address_id"`
	CurrencyID      string          `json:"currency_id"`
	Address         string          `json:"address"`
	Blockchain      Blockchain      `json:"blockchain"`
	Amount          decimal.Decimal `json:"amount"`
	AmountUSD       decimal.Decimal `json:"amount_usd"`
	Dirty           bool            `json:"dirty"`
} // @name WalletWithUSDBalance
