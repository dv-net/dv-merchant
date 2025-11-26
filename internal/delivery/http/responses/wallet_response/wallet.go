package wallet_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/service/wallet"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type WalletAddressResponse struct {
	ID         uuid.UUID `json:"id"`
	WalletID   uuid.UUID `json:"wallet_id"`
	UserID     uuid.UUID `json:"user_id"`
	CurrencyID string    `json:"currency_id"`
	Blockchain string    `json:"blockchain"`
	Address    string    `json:"address"`
	CreatedAt  time.Time `json:"created_at" format:"date-time" swaggertype:"string"`
	UpdatedAt  time.Time `json:"updated_at" format:"date-time" swaggertype:"string"`
	DeletedAt  time.Time `json:"deleted_at" format:"date-time" swaggertype:"string"`
	Dirty      bool      `json:"dirty"`
} //	@name	WalletAddressResponse

type CreateWalletExternalResponse struct {
	ID              uuid.UUID                `json:"id,omitempty"`
	StoreID         uuid.UUID                `json:"store_id"`
	StoreExternalID string                   `json:"store_external_id"`
	AmountUSD       string                   `json:"amount_usd"`
	Address         []*WalletAddressResponse `json:"address"`
	PayURL          string                   `json:"pay_url,omitempty"`
	Rates           map[string]string        `json:"rates"`
	CreatedAt       time.Time                `json:"created_at,omitempty" format:"date-time" swaggertype:"string"`
	UpdatedAt       time.Time                `json:"updated_at,omitempty" format:"date-time" swaggertype:"string"`
} //	@name	CreateWalletExternalResponse

type WalletSeedResponse struct {
	Mnemonic   string `json:"mnemonic"`
	PassPhrase string `json:"pass_phrase"`
} //	@name	WalletSeedResponse

type WalletAddressTotalUSDResponse struct {
	TotalUSD  decimal.Decimal `json:"total_usd"`
	TotalDust decimal.Decimal `json:"total_dust"`
} //	@name	WalletAddressTotalUSDResponse

type ConvertedAddressResponse struct {
	Address *string `json:"address"`
	Legacy  *string `json:"legacy"`
} //	@name	ConvertedAddressResponse

type ExternalProcessingWalletBalanceResponse []*wallet.ProcessingWalletWithAssets //	@name	ExternalProcessingWalletBalanceResponse
