package refund_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/shopspring/decimal"

	"github.com/google/uuid"
)

type CabinetItemResponse struct {
	BlockedTransactionID uuid.UUID           `json:"blocked_transaction_id"`
	TransactionID        uuid.UUID           `json:"transaction_id"`
	TxHash               string              `json:"tx_hash"`
	Blockchain           models.Blockchain   `json:"blockchain"`
	CurrencyID           string              `json:"currency_id"`
	CurrencyCode         string              `json:"currency_code"`
	Amount               decimal.Decimal     `json:"amount"`
	AmountUsd            decimal.NullDecimal `json:"amount_usd"`
	FromAddress          string              `json:"from_address"`
	ToAddress            string              `json:"to_address"`
	RiskLevel            string              `json:"risk_level"`
	Score                decimal.Decimal     `json:"score"`
	CreatedAt            *time.Time          `json:"created_at"`
	RefundStatus         *string             `json:"refund_status,omitempty"`
	DestinationAddress   *string             `json:"destination_address,omitempty"`
} //	@name	CabinetItemResponse
