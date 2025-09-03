package models

import (
	"github.com/shopspring/decimal"
)

type PrefetchWithdrawAddressInfo struct {
	Currency    CurrencyShort   `json:"currency"`
	Amount      decimal.Decimal `json:"amount"`
	AmountUsd   decimal.Decimal `json:"amount_usd"`
	Type        TransferKind    `json:"type"`
	AddressFrom []string        `json:"address_from"`
	AddressTo   []string        `json:"address_to"`
} // @name PrefetchWithdrawAddressInfo
