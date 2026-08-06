package store_request

import (
	"github.com/shopspring/decimal"
)

type UpdateRequest struct {
	Name                     string          `json:"name" validate:"required,min=2,max=32"`
	Site                     *string         `json:"site" validate:""`
	Description              *string         `json:"description"`
	PublicPaymentFormEnabled bool            `json:"public_payment_form_enabled"`
	CurrencyID               string          `json:"currency_id"`
	RateSource               string          `json:"rate_source"`
	ReturnURL                *string         `json:"return_url"`
	SuccessURL               *string         `json:"success_url"`
	RateScale                decimal.Decimal `json:"rate_scale"`
	Status                   bool            `json:"status"`
	MinimalPayment           decimal.Decimal `json:"minimal_payment"`
} //	@name	UpdateStoreRequest
