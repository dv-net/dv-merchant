package user_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

type AuthUserResponse struct {
	ID                uuid.UUID         `json:"id" format:"uuid"`
	Email             string            `json:"email"`
	EmailVerifiedAt   time.Time         `json:"email_verified_at,omitempty" format:"date-time"`
	Location          string            `json:"location"`
	Language          string            `json:"language"`
	RateSource        string            `json:"rate_source"`
	ProcessingOwnerID uuid.UUID         `json:"processing_owner_id" format:"uuid"`
	Roles             []models.UserRole `json:"roles"`
	CreatedAt         time.Time         `json:"created_at,omitempty" format:"date-time"`
	UpdatedAt         time.Time         `json:"updated_at,omitempty" format:"date-time"`
} // @name AuthUserResponse

func (o *AuthUserResponse) Encode(u *models.User, r []models.UserRole) {
	o.ID = u.ID
	o.Email = u.Email
	o.Location = u.Location
	o.Language = u.Language
	o.RateSource = u.RateSource.String()

	if u.ProcessingOwnerID.Valid {
		o.ProcessingOwnerID = u.ProcessingOwnerID.UUID
	}

	if u.EmailVerifiedAt.Valid {
		o.EmailVerifiedAt = u.EmailVerifiedAt.Time
	}

	if u.UpdatedAt.Valid {
		o.UpdatedAt = u.UpdatedAt.Time
	}

	if u.CreatedAt.Valid {
		o.CreatedAt = u.CreatedAt.Time
	}

	o.Roles = r
}
