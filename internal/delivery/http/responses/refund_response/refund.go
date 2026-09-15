package refund_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/service/refund"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// RefundRequestResponse is the admin API shape for a refund request with deposit details.
type RefundRequestResponse struct {
	ID                   uuid.UUID           `json:"id" format:"uuid"`
	BlockedTransactionID uuid.UUID           `json:"blocked_transaction_id" format:"uuid"`
	WalletID             uuid.UUID           `json:"wallet_id" format:"uuid"`
	StoreID              uuid.UUID           `json:"store_id" format:"uuid"`
	TransferID           *uuid.UUID          `json:"transfer_id" format:"uuid"`
	DestinationAddress   string              `json:"destination_address"`
	Status               string              `json:"status"`
	Email                string              `json:"email"`
	ReviewedAt           *time.Time          `json:"reviewed_at" format:"date-time"`
	CreatedAt            time.Time           `json:"created_at" format:"date-time"`
	UpdatedAt            *time.Time          `json:"updated_at" format:"date-time"`
	TransactionID        uuid.UUID           `json:"transaction_id" format:"uuid"`
	TxHash               string              `json:"tx_hash"`
	Amount               decimal.Decimal     `json:"amount"`
	AmountUsd            decimal.NullDecimal `json:"amount_usd"`
	CurrencyID           string              `json:"currency_id"`
	CurrencyCode         string              `json:"currency_code"`
	Blockchain           string              `json:"blockchain"`
	FromAddress          string              `json:"from_address"`
	ToAddress            string              `json:"to_address"`
	RiskLevel            string              `json:"risk_level"`
	Score                decimal.Decimal     `json:"score"`
} //	@name	RefundRequestResponse

func NewRefundRequestResponse(item *refund.RefundRequestDetails) *RefundRequestResponse {
	var transferID *uuid.UUID
	if item.TransferID.Valid {
		id := item.TransferID.UUID
		transferID = &id
	}

	return &RefundRequestResponse{
		ID:                   item.ID,
		BlockedTransactionID: item.BlockedTransactionID,
		WalletID:             item.WalletID,
		StoreID:              item.StoreID,
		TransferID:           transferID,
		DestinationAddress:   item.DestinationAddress,
		Status:               item.Status,
		Email:                item.Email,
		ReviewedAt:           pgtypeutils.DecodeTime(item.ReviewedAt),
		CreatedAt:            item.CreatedAt.Time,
		UpdatedAt:            pgtypeutils.DecodeTime(item.UpdatedAt),
		TransactionID:        item.TransactionID,
		TxHash:               item.TxHash,
		Amount:               item.Amount,
		AmountUsd:            item.AmountUsd,
		CurrencyID:           item.CurrencyID,
		CurrencyCode:         item.CurrencyCode,
		Blockchain:           item.Blockchain.String(),
		FromAddress:          item.FromAddress,
		ToAddress:            item.ToAddress,
		RiskLevel:            item.RiskLevel,
		Score:                item.Score,
	}
}

func NewRefundRequestResponses(items []*refund.RefundRequestDetails) []*RefundRequestResponse {
	res := make([]*RefundRequestResponse, 0, len(items))
	for _, item := range items {
		res = append(res, NewRefundRequestResponse(item))
	}
	return res
}
