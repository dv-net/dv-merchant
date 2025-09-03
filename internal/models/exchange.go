package models

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type WithdrawalFeeType string

func (o WithdrawalFeeType) MarshalBinary() (data []byte, err error) {
	return []byte(o), nil
}
func (o WithdrawalFeeType) String() string { return string(o) }

const (
	WithdrawalFeeTypeFixed WithdrawalFeeType = "fixed"
)

type AddressType string

const (
	DepositAddress  AddressType = "deposit"
	WithdrawAddress AddressType = "withdraw"
)

func (a AddressType) String() string {
	return string(a)
}

type WithdrawalRulesDTO struct {
	Currency            string            `json:"currency" structs:"currency"`
	Chain               string            `json:"chain" structs:"chain"`
	MinDepositAmount    string            `json:"min_deposit_amount" structs:"min_deposit_amount"`
	MinWithdrawAmount   string            `json:"min_withdraw_amount" structs:"min_withdraw_amount"`
	MaxWithdrawAmount   string            `json:"max_withdraw_amount" structs:"max_withdraw_amount"`
	NumOfConfirmations  string            `json:"num_of_confirmations" structs:"num_of_confirmations"`
	WithdrawFeeType     WithdrawalFeeType `json:"withdraw_fee_type" structs:"withdraw_fee_type"`
	WithdrawPrecision   string            `json:"withdraw_precision" structs:"withdraw_precision"`
	WithdrawQuotaPerDay string            `json:"withdraw_quota_per_day" structs:"withdraw_quota_per_day"`
	Fee                 string            `json:"fee" structs:"fee"`
}

type OrderRulesDTO struct {
	Symbol                   string `json:"symbol,omitempty"`
	State                    string `json:"state,omitempty"`
	BaseCurrency             string `json:"base_currency,omitempty"`
	QuoteCurrency            string `json:"quote_currency,omitempty"`
	PricePrecision           int    `json:"price_precision,omitempty"`
	AmountPrecision          int    `json:"amount_precision,omitempty"`
	ValuePrecision           int    `json:"value_precision,omitempty"`
	MinOrderAmount           string `json:"min_order_amount,omitempty"`
	MaxOrderAmount           string `json:"max_order_amount,omitempty"`
	MinOrderValue            string `json:"min_order_value,omitempty"`
	SellMarketMinOrderAmount string `json:"sell_market_min_order_amount,omitempty"`
	SellMarketMaxOrderAmount string `json:"sell_market_max_order_amount,omitempty"`
	BuyMarketMaxOrderValue   string `json:"buy_market_max_order_value,omitempty"`
}

type OrderDetailsDTO struct {
	State      ExchangeOrderStatus `json:"state,omitempty"`
	Amount     decimal.Decimal     `json:"amount"`
	AmountUSD  decimal.Decimal     `json:"amount_usd"`
	FailReason string              `json:"fail_reason,omitempty"`
}

type AccountBalanceDTO struct {
	Currency  string          `json:"currency"`
	Type      string          `json:"type"`
	Amount    decimal.Decimal `json:"amount"`
	AmountUSD decimal.Decimal `json:"amount_usd"`
}

type ExchangeSymbolDTO struct {
	Symbol      string `json:"symbol"`
	DisplayName string `json:"display_name"`
	BaseSymbol  string `json:"base_symbol"`
	QuoteSymbol string `json:"quote_symbol"`
	Type        string `json:"type"`
}

type DepositAddressDTO struct {
	Address          string      `json:"address"`
	Currency         string      `json:"currency"`
	InternalCurrency string      `json:"internal_currency"`
	Chain            string      `json:"chain"`
	AddressType      AddressType `json:"address_type"`
	PaymentTag       string      `json:"payment_tag,omitempty"`
}

type WithdrawalAddressDTO struct {
	Address  string `json:"address"`
	Currency string `json:"currency"`
}

type OrderSide string

func (o OrderSide) String() string { return string(o) }

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

type OrderType string

func (o OrderType) String() string { return string(o) }

const (
	OrderTypeMarket OrderType = ""
	OrderTypeLimit  OrderType = ""
)

type ExchangeGroup struct {
	Slug      string              `json:"slug"`
	Name      string              `json:"name"`
	Addresses []DepositAddressDTO `json:"addresses"`
}

type ExchangeWithdrawalDTO struct {
	ExternalOrderID string `json:"external_order_id"`
	InternalOrderID string `json:"internal_order_id"`
	RetryReason     string `json:"retry_reason"`
}

type WithdrawalHistoryStatus string

func (o WithdrawalHistoryStatus) String() string { return string(o) }

const (
	WithdrawalHistoryStatusNew        WithdrawalHistoryStatus = "new"
	WithdrawalHistoryStatusFailed     WithdrawalHistoryStatus = "failed"
	WithdrawalHistoryStatusInProgress WithdrawalHistoryStatus = "in_progress"
	WithdrawalHistoryStatusCompleted  WithdrawalHistoryStatus = "completed"
	WithdrawalHistoryStatusCanceled   WithdrawalHistoryStatus = "canceled"
	WithdrawalHistoryStatusRecovery   WithdrawalHistoryStatus = "recovery"
)

type WithdrawalStatusDTO struct {
	ID           string          `json:"id"`
	Status       string          `json:"status"`
	TxHash       string          `json:"tx_hash"`
	NativeAmount decimal.Decimal `json:"native_amount"`
	ErrorMessage string          `json:"error_message,omitempty"`
}

type CreateWithdrawalOrderParams struct {
	RecordID            *uuid.UUID
	Currency            string
	Chain               string
	NativeAmount        decimal.Decimal
	FiatAmount          decimal.Decimal
	Address             string
	Fee                 decimal.Decimal
	WithdrawalPrecision int
	MinWithdrawal       decimal.Decimal
}

type GetWithdrawalByIDParams struct {
	ExternalOrderID *string
	ClientOrderID   *string
}

type GetOrderByIDParams struct {
	InstrumentID    *string
	ExternalOrderID *string
	ClientOrderID   *string
	OrderSide       string
	InternalOrder   *ExchangeOrder
}

type ExchangeOrderDTO struct {
	ClientOrderID   string          `json:"client_order_id"`
	ExchangeOrderID string          `json:"exchange_order_id"`
	Amount          decimal.Decimal `json:"amount"`
}

type ExchangeWithdrawalHistoryDTO struct {
	ID              uuid.UUID               `db:"id" json:"id"`
	UserID          uuid.UUID               `db:"user_id" json:"user_id"`
	ExchangeID      uuid.UUID               `db:"exchange_id" json:"exchange_id"`
	ExchangeOrderID pgtype.Text             `db:"exchange_order_id" json:"exchange_order_id"`
	Address         string                  `db:"address" json:"address"`
	MinAmount       decimal.Decimal         `db:"min_amount" json:"min_amount"`
	NativeAmount    decimal.NullDecimal     `db:"native_amount" json:"native_amount"`
	FiatAmount      decimal.NullDecimal     `db:"fiat_amount" json:"fiat_amount"`
	Currency        string                  `db:"currency" json:"currency"`
	Chain           string                  `db:"chain" json:"chain"`
	Status          WithdrawalHistoryStatus `db:"status" json:"status"`
	Txid            pgtype.Text             `db:"txid" json:"txid"`
	CreatedAt       pgtype.Timestamp        `db:"created_at" json:"created_at"`
	UpdatedAt       pgtype.Timestamp        `db:"updated_at" json:"updated_at"`
	Slug            ExchangeSlug            `db:"slug" json:"slug"`
	FailReason      pgtype.Text             `db:"fail_reason" json:"fail_reason"`
}

type ExchangeOrderStatus string

func (o ExchangeOrderStatus) String() string { return string(o) }

const (
	ExchangeOrderStatusNew        ExchangeOrderStatus = "new"
	ExchangeOrderStatusInProgress ExchangeOrderStatus = "in_progress"
	ExchangeOrderStatusCompleted  ExchangeOrderStatus = "completed"
	ExchangeOrderStatusFailed     ExchangeOrderStatus = "failed"
)

type ExchangeWithdrawalState string

func (o ExchangeWithdrawalState) String() string { return string(o) }

func (o ExchangeWithdrawalState) Invert() ExchangeWithdrawalState {
	if o == ExchangeWithdrawalStateEnabled {
		return ExchangeWithdrawalStateDisabled
	}
	return ExchangeWithdrawalStateEnabled
}

const (
	ExchangeWithdrawalStateDisabled ExchangeWithdrawalState = "disabled"
	ExchangeWithdrawalStateEnabled  ExchangeWithdrawalState = "enabled"
)

type ExchangeSwapState string

func (o ExchangeSwapState) String() string { return string(o) }

func (o ExchangeSwapState) Invert() ExchangeSwapState {
	if o == ExchangeSwapStateEnabled {
		return ExchangeSwapStateDisabled
	}
	return ExchangeSwapStateEnabled
}

const (
	ExchangeSwapStateDisabled ExchangeSwapState = "disabled"
	ExchangeSwapStateEnabled  ExchangeSwapState = "enabled"
)
