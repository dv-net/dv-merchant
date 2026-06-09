package repo_stores

import "github.com/dv-net/dv-merchant/internal/models"

func (r *GetStoreByWalletAddressRow) ToStore() *models.Store {
	return &models.Store{
		ID:                       r.ID,
		UserID:                   r.UserID,
		Name:                     r.Name,
		Site:                     r.Site,
		CurrencyID:               r.CurrencyID,
		RateSource:               r.RateSource,
		ReturnUrl:                r.ReturnUrl,
		SuccessUrl:               r.SuccessUrl,
		RateScale:                r.RateScale,
		Status:                   r.Status,
		MinimalPayment:           r.MinimalPayment,
		CreatedAt:                r.CreatedAt,
		UpdatedAt:                r.UpdatedAt,
		DeletedAt:                r.DeletedAt,
		PublicPaymentFormEnabled: r.PublicPaymentFormEnabled,
	}
}
