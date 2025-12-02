package store

import (
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

type CreateStore struct {
	Name string
	Site *string
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
