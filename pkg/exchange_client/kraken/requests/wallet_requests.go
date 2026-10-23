package requests

import (
	"time"

	"github.com/shopspring/decimal"
)

type DepositMethodsRequest struct {
	Asset  string
	AClass string
}

type DepositAddressesRequest struct {
	Asset  string
	Method string // e.g. "Tether USD (TRC20)", see DepositMethods
	New    bool   // generate a new address
	AClass string
}

type WithdrawMethodsRequest struct {
	Asset   string
	AClass  string
	Network string
}

type WithdrawAddressesRequest struct {
	Asset    string
	AClass   string
	Method   string
	Key      string
	Verified *bool
}

type WithdrawInfoRequest struct {
	Asset  string
	Key    string // name of the withdrawal address saved in Kraken
	Amount decimal.Decimal
}

type WithdrawRequest struct {
	Asset  string
	Key    string // name of the withdrawal address saved in Kraken
	Amount decimal.Decimal
	// Kraken rejects the withdrawal if Key does not point to this address.
	Address string
	// Kraken rejects the withdrawal if the fee is higher. Zero means no limit.
	MaxFee decimal.Decimal
	AClass string
}

type WithdrawStatusRequest struct {
	Asset  string
	AClass string
	Method string
	Start  time.Time
	End    time.Time
}
