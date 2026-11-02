package refund_response

import (
	"github.com/dv-net/dv-merchant/internal/service/refund"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// RefundEstimateResponse is the expected result of a refund, the fee may change by the time of sending
type RefundEstimateResponse struct {
	BlockedTransactionID uuid.UUID           `json:"blocked_transaction_id" format:"uuid"`
	Amount               decimal.Decimal     `json:"amount"`
	CurrencyID           string              `json:"currency_id"`
	EstimatedFee         decimal.NullDecimal `json:"estimated_fee" swaggertype:"string"`
	FeeCurrencyID        string              `json:"fee_currency_id"`
	FeeFromAmount        bool                `json:"fee_from_amount"`
	AmountToReceive      decimal.NullDecimal `json:"amount_to_receive" swaggertype:"string"`
	FeeCovered           bool                `json:"fee_covered"`
	FeeBalance           decimal.NullDecimal `json:"fee_balance" swaggertype:"string"`
	FeeShortage          decimal.NullDecimal `json:"fee_shortage" swaggertype:"string"`
	ActivationRequired   bool                `json:"activation_required"`
	RequiresReview       bool                `json:"requires_review"`
} //	@name	RefundEstimateResponse

func NewRefundEstimateResponse(e *refund.EstimateDTO) *RefundEstimateResponse {
	return &RefundEstimateResponse{
		BlockedTransactionID: e.BlockedTransactionID,
		Amount:               e.Amount,
		CurrencyID:           e.CurrencyID,
		EstimatedFee:         e.EstimatedFee,
		FeeCurrencyID:        e.FeeCurrencyID,
		FeeFromAmount:        e.FeeFromAmount,
		AmountToReceive:      e.AmountToReceive,
		FeeCovered:           e.FeeCovered,
		FeeBalance:           e.FeeBalance,
		FeeShortage:          e.FeeShortage,
		ActivationRequired:   e.ActivationRequired,
		RequiresReview:       e.RequiresReview,
	}
}
