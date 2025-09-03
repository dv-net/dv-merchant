package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/receipt_response"
	"github.com/dv-net/dv-merchant/internal/models"
)

func FromReceiptModelToResponse(r *models.Receipt) *receipt_response.ReceiptResponse {
	return &receipt_response.ReceiptResponse{
		ID:         r.ID.String(),
		Status:     r.Status,
		StoreID:    r.StoreID.String(),
		CurrencyID: r.CurrencyID,
		Amount:     r.Amount,
		WalletID:   r.WalletID.UUID.String(),
		CreatedAt:  r.CreatedAt.Time,
		UpdatedAt:  r.UpdatedAt.Time,
	}
}

func FromReceiptModelToResponses(models ...*models.Receipt) []*receipt_response.ReceiptResponse {
	res := make([]*receipt_response.ReceiptResponse, 0, len(models))
	for _, model := range models {
		res = append(res, FromReceiptModelToResponse(model))
	}
	return res
}
