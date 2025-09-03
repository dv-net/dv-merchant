package converters

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/transfer_response"
	"github.com/dv-net/dv-merchant/internal/models"
)

func FromTransferPrefetchModelToResponse(o *models.PrefetchWithdrawAddressInfo) *transfer_response.GetPrefetchDataResponse {
	res := &transfer_response.GetPrefetchDataResponse{
		Currency:      o.Currency,
		Amount:        o.Amount,
		AmountUsd:     o.AmountUsd,
		Type:          o.Type,
		FromAddresses: o.AddressFrom,
		ToAddresses:   o.AddressTo,
	}
	return res
}

func FromTransferPrefetchModelToResponses(o ...*models.PrefetchWithdrawAddressInfo) []*transfer_response.GetPrefetchDataResponse {
	res := make([]*transfer_response.GetPrefetchDataResponse, 0, len(o))
	for _, v := range o {
		res = append(res, FromTransferPrefetchModelToResponse(v))
	}
	return res
}

func FromTransferModelToResponse(o *models.Transfer) *transfer_response.GetTransferResponse {
	var createdAt *time.Time
	if o.CreatedAt.Valid {
		createdAt = &o.CreatedAt.Time
	}

	var updatedAt *time.Time
	if o.UpdatedAt.Valid {
		updatedAt = &o.UpdatedAt.Time
	}

	return &transfer_response.GetTransferResponse{
		ID:            o.ID.String(),
		Number:        o.Number,
		UserID:        o.UserID.String(),
		Kind:          o.Kind,
		CurrencyID:    o.CurrencyID,
		Status:        o.Status,
		Step:          o.Step,
		FromAddresses: o.FromAddresses,
		ToAddresses:   o.ToAddresses,
		TxHash:        o.TxHash,
		Amount:        o.Amount,
		AmountUsd:     o.AmountUsd,
		Message:       o.Message,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

func FromTransferModelToResponses(o ...*models.Transfer) []*transfer_response.GetTransferResponse {
	res := make([]*transfer_response.GetTransferResponse, 0, len(o))
	for _, v := range o {
		res = append(res, FromTransferModelToResponse(v))
	}
	return res
}
