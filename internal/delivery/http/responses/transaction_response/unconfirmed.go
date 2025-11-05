package transaction_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UnconfirmedTransactionResponse struct {
	ID               uuid.UUID           `json:"id" format:"uuid"`
	CurrencyID       string              `json:"currency_id"`
	Blockchain       string              `json:"blockchain"`
	TxHash           string              `json:"tx_hash"`
	BcUniqKey        *string             `json:"bc_uniq_key"`
	Type             string              `json:"type"`
	FromAddress      string              `json:"from_address"`
	ToAddress        string              `json:"to_address"`
	Amount           decimal.Decimal     `json:"amount"`
	AmountUsd        decimal.NullDecimal `json:"amount_usd"`
	NetworkCreatedAt time.Time           `json:"network_created_at" format:"date-time"`
	CreatedAt        time.Time           `json:"created_at" format:"date-time"`
	UpdatedAt        *time.Time          `json:"updated_at,omitempty" format:"date-time"`
} // @name UnconfirmedTransactionResponse

func NewFromUnconfirmedTransactionModel(tx *models.UnconfirmedTransaction) *UnconfirmedTransactionResponse {
	return &UnconfirmedTransactionResponse{
		ID:               tx.ID,
		CurrencyID:       tx.CurrencyID,
		Blockchain:       tx.Blockchain.String(),
		TxHash:           tx.TxHash,
		BcUniqKey:        tx.BcUniqKey,
		Type:             tx.Type.String(),
		FromAddress:      tx.FromAddress,
		ToAddress:        tx.ToAddress,
		Amount:           tx.Amount,
		AmountUsd:        tx.AmountUsd,
		NetworkCreatedAt: tx.NetworkCreatedAt.Time,
		CreatedAt:        tx.CreatedAt.Time,
		UpdatedAt:        pgtypeutils.DecodeTime(tx.UpdatedAt),
	}
}

func NewFromUnconfirmedTransactionModels(cur []*models.UnconfirmedTransaction) []*UnconfirmedTransactionResponse {
	responses := make([]*UnconfirmedTransactionResponse, 0, len(cur))
	for _, c := range cur {
		responses = append(responses, NewFromUnconfirmedTransactionModel(c))
	}
	return responses
}
