package public_request

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/shopspring/decimal"

	"github.com/google/uuid"
)

type RefundLookupRequest struct {
	WalletID uuid.UUID `json:"wallet_id" validate:"required"`
	StoreID  uuid.UUID `json:"store_id" validate:"required"`
	Email    string    `json:"email" validate:"required,email"`
} //	@name	RefundLookupRequest

type RefundVerifyRequest struct {
	WalletID uuid.UUID `json:"wallet_id" validate:"required"`
	StoreID  uuid.UUID `json:"store_id" validate:"required"`
	Email    string    `json:"email" validate:"required,email"`
	Code     string    `json:"code" validate:"required,len=6,alphanum"`
} //	@name	RefundVerifyRequest

type RefundVerifyResponse struct {
	Token string `json:"token"`
} //	@name	RefundVerifyResponse

type RefundClaimRequest struct {
	DestinationAddress string `json:"destination_address" validate:"required"`
} //	@name	RefundClaimRequest

type CabinetItemResponse struct {
	BlockedTransactionID uuid.UUID         `json:"blocked_transaction_id"`
	TransactionID        uuid.UUID         `json:"transaction_id"`
	TxHash               string            `json:"tx_hash"`
	Blockchain           models.Blockchain `json:"blockchain"`
	CurrencyID           string            `json:"currency_id"`
	RiskLevel            string            `json:"risk_level"`
	Score                decimal.Decimal   `json:"score"`
	CreatedAt            *time.Time        `json:"created_at"`
	RefundStatus         *string           `json:"refund_status,omitempty"`
	DestinationAddress   *string           `json:"destination_address,omitempty"`
} //	@name	CabinetItemResponse
