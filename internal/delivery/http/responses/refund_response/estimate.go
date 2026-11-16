package refund_response

import (
	"github.com/dv-net/dv-merchant/internal/service/refund"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// RefundEstimateResponse is the expected result of a refund, the fee may change by the time of sending
type RefundEstimateResponse struct {
	BlockedTransactionID uuid.UUID           `json:"blocked_transaction_id" format:"uuid"`
	DestinationAddress   string              `json:"destination_address"`
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
		DestinationAddress:   e.DestinationAddress,
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

// ExternalRefundEstimateResponse is the expected result of a refund requested through the external api.
// CodeRequired reports that the address is not a deposit sender: a confirmation code must be sent
// to the wallet email (send-code) and passed on the refund creation
//
// FromAddress is the deposit address the refund is sent from. The network fee is paid from it:
// when fee_covered is false, top it up with fee_shortage of fee_currency_id
type ExternalRefundEstimateResponse struct {
	*RefundEstimateResponse
	FromAddress  string `json:"from_address"`
	CodeRequired bool   `json:"code_required"`
} //	@name	ExternalRefundEstimateResponse

func NewExternalRefundEstimateResponse(e *refund.EstimateDTO) *ExternalRefundEstimateResponse {
	return &ExternalRefundEstimateResponse{
		RefundEstimateResponse: NewRefundEstimateResponse(e),
		FromAddress:            e.FromAddress,
		CodeRequired:           e.RequiresReview,
	}
}
