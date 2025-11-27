package user_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetUserInfoResponse struct {
	ID                    uuid.UUID         `json:"id" format:"uuid"`
	Email                 string            `json:"email" validate:"required,email" format:"email"`
	EmailVerifiedAt       *time.Time        `json:"email_verified_at" format:"date-time"`
	Location              string            `json:"location" validate:"required,timezone"`
	Language              string            `json:"language"`
	RateSource            string            `json:"rate_source"`
	RateScale             decimal.Decimal   `json:"rate_scale"`
	ProcessingOwnerID     uuid.NullUUID     `json:"processing_owner_id" format:"uuid"`
	Roles                 []models.UserRole `json:"roles"`
	CreatedAt             time.Time         `json:"created_at" format:"date-time"`
	UpdatedAt             time.Time         `json:"updated_at" format:"date-time"`
	QuickStartGuideStatus string            `json:"quick_start_guide_status"`
} //	@name	GetUserInfoResponse

type TgLinkResponse struct {
	Link string `json:"link"`
} //	@name	TgLinkResponse
