package store_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type StoreResponse struct {
	ID                       uuid.UUID       `json:"id" format:"uuid"`
	UserID                   uuid.UUID       `json:"user_id" format:"uuid"`
	Name                     string          `json:"name"`
	Site                     *string         `json:"site" format:"uri"`
	CurrencyID               string          `json:"currency_id"`
	RateSource               string          `json:"rate_source"`
	ReturnURL                *string         `json:"return_url" format:"uri"`
	SuccessURL               *string         `json:"success_url" format:"uri"`
	RateScale                decimal.Decimal `json:"rate_scale"`
	Status                   bool            `json:"status"`
	MinimalPayment           decimal.Decimal `json:"minimal_payment"`
	PublicPaymentFormEnabled bool            `json:"public_payment_form_enabled"`
	CreatedAt                time.Time       `json:"created_at" format:"date-time"`
} //	@name	StoreResponse

type StoreWithTransactionsResponse struct {
	ID             uuid.UUID       `json:"id" format:"uuid"`
	UserID         uuid.UUID       `json:"user_id" format:"uuid"`
	Name           string          `json:"name"`
	Site           *string         `json:"site" format:"uri"`
	CurrencyID     string          `json:"currency_id"`
	RateSource     string          `json:"rate_source"`
	ReturnURL      *string         `json:"return_url" format:"uri"`
	SuccessURL     *string         `json:"success_url" format:"uri"`
	RateScale      decimal.Decimal `json:"rate_scale"`
	Status         bool            `json:"status"`
	MinimalPayment decimal.Decimal `json:"minimal_payment"`
	CreatedAt      time.Time       `json:"created_at" format:"date-time"`
	// StoreResponse
	TransactionCount int `json:"transaction_count"`
} //	@name	StoreWithTransactionsResponse

type StoreAPIKeyResponse struct {
	ID      string `json:"id" format:"uuid"`
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
} //	@name	StoreAPIKeyResponse

type StoreSecretResponse struct {
	Secret string `json:"secret"`
} //	@name	StoreSecretResponse

type StoreWebhookResponse struct {
	ID string `json:"id" format:"uuid"`
	// StoreID   string    `json:"store_id" format:"uuid"`
	URL     string   `json:"url" format:"uri"`
	Enabled bool     `json:"enabled"`
	Events  []string `json:"events" enums:"PaymentReceived,PaymentNotConfirmed"`
	// CreatedAt time.Time `json:"created_at" format:"date-time"`
} //	@name	StoreWebhookResponse

type StoreTransactionResponse struct {
	ID                 string                  `json:"id" format:"uuid"`
	UserID             string                  `json:"user_id" format:"uuid"`
	StoreID            string                  `json:"store_id" format:"uuid"`
	ReceiptID          *string                 `json:"receipt_id" format:"uuid"`
	WalletID           *string                 `json:"wallet_id" format:"uuid"`
	CurrencyID         string                  `json:"currency_id"`
	Blockchain         string                  `json:"blockchain"`
	TxHash             string                  `json:"tx_hash"`
	BcUniqKey          *string                 `json:"bc_uniq_key"`
	Type               models.TransactionsType `json:"type"`
	FromAddress        string                  `json:"from_address"`
	ToAddress          string                  `json:"to_address"`
	Amount             decimal.Decimal         `json:"amount"`
	AmountUsd          decimal.NullDecimal     `json:"amount_usd"`
	Fee                decimal.Decimal         `json:"fee"`
	WithdrawalIsManual bool                    `json:"withdrawal_is_manual"`
	NetworkCreatedAt   time.Time               `json:"network_created_at" format:"date-time"`
} //	@name	StoreTransactionResponse

type StoreCurrencyResponse []string //	@name	StoreCurrencyResponse

type StoreWhitelistResponse []string //	@name	StoreWhitelistResponse
