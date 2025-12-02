package admin_request

import (
	"github.com/google/uuid"
)

type BanUserRequest struct {
	UserID uuid.UUID `json:"user_id" query:"user_id" format:"uuid"`
} //	@name	BanUserRequest
