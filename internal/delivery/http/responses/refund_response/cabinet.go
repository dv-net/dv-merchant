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
	// BlockReasons are the AML rules which blocked the transaction, empty for transactions blocked before it was recorded
	BlockReasons       []CabinetBlockReasonResponse `json:"block_reasons"`
	CreatedAt          *time.Time                   `json:"created_at"`
	RefundStatus       *string                      `json:"refund_status,omitempty"`
	DestinationAddress *string                      `json:"destination_address,omitempty"`
	// RefundExecuteAfter is set for scheduled refunds to the deposit sender
	RefundExecuteAfter *time.Time `json:"refund_execute_after,omitempty" format:"date-time"`
} //	@name	CabinetItemResponse

// CabinetBlockReasonResponse is an AML rule which blocked the transaction. The rule's threshold
// is not shown to the payer.
type CabinetBlockReasonResponse struct {
	RiskType string          `json:"risk_type"`
	Value    decimal.Decimal `json:"value"`
} //	@name	CabinetBlockReasonResponse

func NewCabinetBlockReasonsResponse(reasons models.AmlBlockReasons) []CabinetBlockReasonResponse {
	res := make([]CabinetBlockReasonResponse, 0, len(reasons))
	for _, r := range reasons {
		res = append(res, CabinetBlockReasonResponse{RiskType: r.RiskType, Value: r.Value})
	}
	return res
}
