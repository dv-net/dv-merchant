package store_request

import (
	"github.com/shopspring/decimal"
)

type UpdateRequest struct {
	Name                     string          `db:"name" json:"name" validate:"required,min=2,max=32"`
	Site                     *string         `db:"site" json:"site" validate:""`
	PublicPaymentFormEnabled bool            `db:"public_payment_form_enabled" json:"public_payment_form_enabled"`
	CurrencyID               string          `db:"currency_id" json:"currency_id"`
	RateSource               string          `db:"rate_source" json:"rate_source"`
	ReturnURL                *string         `db:"return_url" json:"return_url"`
	SuccessURL               *string         `db:"success_url" json:"success_url"`
	RateScale                decimal.Decimal `db:"rate_scale" json:"rate_scale"`
	Status                   bool            `db:"status" json:"status"`
	MinimalPayment           decimal.Decimal `db:"minimal_payment" json:"minimal_payment"`
} //	@name	UpdateStoreRequest
