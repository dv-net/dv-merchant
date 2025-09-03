package exchange_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/shopspring/decimal"
)

type ExchangeListResponse struct {
	Exchanges       []ExchangeData `json:"exchanges"`
	CurrentExchange *string        `json:"current_exchange"`
	SwapState       *string        `json:"swap_state"`
	WithdrawalState *string        `json:"withdrawal_state"`
} // @name ExchangeWithKeysResponse

type ExchangeData struct {
	Exchange            string        `json:"exchange"`
	Slug                string        `json:"slug"`
	ExchangeConnectedAt time.Time     `json:"exchange_connected_at"`
	Keys                []ExchangeKey `json:"keys"`
}

type ExchangeKey struct {
	Name  string  `json:"name"`
	Value *string `json:"value"`
} // @name ExchangeKeyData

type ExchangeAsset struct {
	Currency  string          `json:"currency"`
	Amount    decimal.Decimal `json:"amount"`
	AmountUSD decimal.Decimal `json:"amount_usd"`
} // @name ExchangeAsset

type ExchangeBalanceResponse struct {
	Balances []ExchangeAsset `json:"balances"`
	TotalUSD decimal.Decimal `json:"total_usd"`
} // @name ExchangeBalance

type ExchangeTestConnectionResponse struct {
	Exchange     string `json:"exchange"`
	ErrorMessage string `json:"error_message"`
} // @name ExchangeTestConnectionResponse

type ExchangeUserPairResponse struct {
	DisplayName string `json:"display_name"`
} // @name ExchangeUserPairResponse

type ExternalExchangeBalanceResponse ExchangeBalanceResponse // @name ExternalExchangeBalanceResponse

type ExchangeCreateWithdrawalResponse struct {
	InternalOrderID string `json:"internal_order_id"`
} // @name ExchangeCreateWithdrawalResponse

type ExchangeWithdrawalHistoryResponse struct {
	ID           string    `json:"id"`
	CurrencyID   string    `json:"currency_id"`
	ExchangeID   string    `json:"exchange_id"`
	ExchangeSlug string    `json:"exchange_slug"`
	Chain        string    `json:"chain"`
	AmountNative *string   `json:"amount_native"`
	AmountUSD    *string   `json:"amount_usd"`
	Address      string    `json:"address"`
	TxID         *string   `json:"tx_id"`
	Status       string    `json:"status"`
	FailReason   *string   `json:"fail_reason"`
	CreatedAt    time.Time `json:"created_at" format:"date-time"`
} // @name ExchangeWithdrawalHistoryResponse

type ExchangeOrderHistoryResponse struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	ExchangeID      string    `json:"exchange_id"`
	ExchangeSlug    string    `json:"exchange_slug"`
	ExchangeOrderID string    `json:"exchange_order_id"`
	ClientOrderID   string    `json:"client_order_id"`
	Symbol          string    `json:"symbol"`
	Side            string    `json:"side"`
	Amount          string    `json:"amount"`
	AmountUsd       string    `json:"amount_usd"`
	Status          string    `json:"status"`
	FailReason      string    `json:"fail_reason"`
	CreatedAt       time.Time `json:"created_at" format:"date-time"`
} // @name ExchangeOrderHistoryResponse

type ExchangeWithdrawalRulesResponse struct {
	Currency            string  `json:"currency" structs:"currency"`
	Chain               string  `json:"chain" structs:"chain"`
	MinDepositAmount    string  `json:"min_deposit_amount" structs:"min_deposit_amount"`
	MinWithdrawAmount   string  `json:"min_withdraw_amount" structs:"min_withdraw_amount"`
	MaxWithdrawAmount   *string `json:"max_withdraw_amount" structs:"max_withdraw_amount"`
	NumOfConfirmations  string  `json:"num_of_confirmations" structs:"num_of_confirmations"`
	WithdrawFeeType     *string `json:"withdraw_fee_type" structs:"withdraw_fee_type"`
	WithdrawPrecision   string  `json:"withdraw_precision" structs:"withdraw_precision"`
	WithdrawQuotaPerDay *string `json:"withdraw_quota_per_day" structs:"withdraw_quota_per_day"`
	Fee                 *string `json:"fee" structs:"fee"`
} // @ExchangeWithdrawalRulesResponse

type ExchangeWithdrawalSettingResponse struct {
	ID         string          `json:"id"`
	CurrencyID string          `json:"currency_id"`
	Chain      string          `json:"chain"`
	Address    string          `json:"address"`
	MinAmount  decimal.Decimal `json:"min_amount"`
	Enabled    bool            `json:"enabled"`
	CreatedAt  time.Time       `json:"created_at"`
} // @name ExchangeWithdrawalSettingResponse

type DepositUpdateResponse struct {
	Address          string             `json:"address"`
	Currency         string             `json:"currency"`
	Chain            string             `json:"chain"`
	AddressType      models.AddressType `json:"address_type"`
	MinDepositAmount string             `json:"min_deposit_amount"`
} // @name DepositUpdateResponse

type GetDepositAddressesResponse struct {
	Slug      string                  `json:"slug"`
	Name      string                  `json:"name"`
	Addresses []DepositUpdateResponse `json:"addresses"`
} // @name GetDepositAddressesResponse
