package public_request

import (
	"github.com/google/uuid"
)

type RefundLookupRequest struct {
	WalletID uuid.UUID `json:"wallet_id" validate:"required"`
	StoreID  uuid.UUID `json:"store_id" validate:"required"`
	Email    string    `json:"email" validate:"required,email"`
} //	@name	RefundLookupRequest

type RefundVerifyRequest struct {
	WalletID uuid.UUID `json:"wallet_id" validate:"required"`
	StoreID  uuid.UUID `json:"store_id" validate:"required"`
	Email    string    `json:"email" validate:"required,email"`
	Code     string    `json:"code" validate:"required,len=6,alphanum"`
} //	@name	RefundVerifyRequest

type RefundVerifyResponse struct {
	Token string `json:"token"`
} //	@name	RefundVerifyResponse

type RefundClaimRequest struct {
	DestinationAddress string `json:"destination_address" validate:"required"`
} //	@name	RefundClaimRequest
