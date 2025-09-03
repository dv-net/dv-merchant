package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/store_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/util"
)

func FromStoreModelToResponse(o *models.Store) *store_response.StoreResponse {
	res := &store_response.StoreResponse{
		ID:                       o.ID,
		UserID:                   o.UserID,
		Name:                     o.Name,
		Site:                     o.Site,
		CurrencyID:               o.CurrencyID,
		RateSource:               o.RateSource.String(),
		RateScale:                o.RateScale,
		Status:                   o.Status,
		MinimalPayment:           o.MinimalPayment,
		PublicPaymentFormEnabled: o.PublicPaymentFormEnabled,
		CreatedAt:                o.CreatedAt.Time,
		ReturnURL:                o.ReturnUrl,
		SuccessURL:               o.SuccessUrl,
	}
	return res
}

func FromStoreModelToResponses(models ...*models.Store) []*store_response.StoreResponse {
	res := make([]*store_response.StoreResponse, 0, len(models))
	for _, model := range models {
		res = append(res, FromStoreModelToResponse(model))
	}
	return res
}

func FromStoreTransactionModelToResponse(t *models.Transaction) *store_response.StoreTransactionResponse {
	res := &store_response.StoreTransactionResponse{
		ID:                 t.ID.String(),
		UserID:             t.UserID.String(),
		StoreID:            t.StoreID.UUID.String(),
		CurrencyID:         t.CurrencyID,
		Blockchain:         t.Blockchain,
		TxHash:             t.TxHash,
		Type:               t.Type,
		FromAddress:        t.FromAddress,
		ToAddress:          t.ToAddress,
		Amount:             t.Amount,
		AmountUsd:          t.AmountUsd,
		Fee:                t.Fee,
		WithdrawalIsManual: t.WithdrawalIsManual,
		NetworkCreatedAt:   t.NetworkCreatedAt.Time,
		BcUniqKey:          t.BcUniqKey,
	}

	if t.ReceiptID.Valid {
		res.ReceiptID = util.Pointer(t.ReceiptID.UUID.String())
	}

	if t.WalletID.Valid {
		res.WalletID = util.Pointer(t.WalletID.UUID.String())
	}

	return res
}

func FromStoreTransactionModelToResponses(transactions ...*models.Transaction) []*store_response.StoreTransactionResponse {
	res := make([]*store_response.StoreTransactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		res = append(res, FromStoreTransactionModelToResponse(transaction))
	}
	return res
}

func FromStoreWhitelistModelToResponse(o *models.StoreWhitelist) *store_response.StoreWhitelistResponse {
	return &store_response.StoreWhitelistResponse{o.Ip}
}

func FromStoreWhitelistModelToResponses(m ...*models.StoreWhitelist) store_response.StoreWhitelistResponse {
	res := make(store_response.StoreWhitelistResponse, 0, len(m))
	for _, model := range m {
		res = append(res, model.Ip)
	}
	return res
}
