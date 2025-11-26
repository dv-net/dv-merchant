package wallet_request

import "github.com/shopspring/decimal"

type CreateRequest struct {
	StoreID         string          `json:"store_id,omitempty" validate:"required,uuid" format:"uuid"`
	StoreExternalID string          `json:"store_external_id" validate:"required" format:"uuid"`
	Amount          decimal.Decimal `json:"amount" query:"amount" validate:"omitempty"`
	Currency        *string         `json:"currency,omitempty" query:"currency" validate:"omitempty"`
	Email           *string         `json:"email,omitempty" query:"email" validate:"omitempty,email" format:"email"`
	Locale          *string         `json:"locale,omitempty" query:"locale" validate:"omitempty"`
} //	@name	CreateWalletRequest

type ExternalCreateRequest struct {
	StoreExternalID string          `json:"store_external_id" query:"store_external_id" validate:"required" format:"uuid"`
	IP              *string         `json:"ip,omitempty" query:"ip" validate:"omitempty,ip" format:"ipv4"`
	Email           *string         `json:"email,omitempty" query:"email" validate:"omitempty,email" format:"email"`
	Amount          decimal.Decimal `json:"amount" query:"amount" validate:"omitempty"`
	Currency        *string         `json:"currency,omitempty" query:"currency" validate:"omitempty"`
	Locale          *string         `json:"locale,omitempty" query:"locale" validate:"omitempty"`
} //	@name	CreateWalletExternalRequest
