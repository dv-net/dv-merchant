package store

import (
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/shopspring/decimal"

	"github.com/google/uuid"
)

type CreateStore struct {
	Name        string
	Site        *string
	Description *string
}

type UpdateStore struct {
	Name                     string
	Site                     *string
	Description              *string
	PublicPaymentFormEnabled bool
	CurrencyID               string
	RateSource               string
	ReturnURL                *string
	SuccessURL               *string
	RateScale                decimal.Decimal
	Status                   bool
	MinimalPayment           decimal.Decimal
}

type CurrencyRate struct {
	Code       string `json:"code"`
	RateSource string `json:"rate_source"`
	Rate       string `json:"rate"`
} //	@name	CurrencyRate

type ArchiveStoreDTO struct {
	OTP     string       `json:"otp"`
	User    *models.User `json:"user"`
	StoreID uuid.UUID    `json:"store_id"`
}
