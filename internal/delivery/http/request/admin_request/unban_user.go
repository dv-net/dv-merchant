package admin_request

import (
	"github.com/google/uuid"
)

type UnbanUserRequest struct {
	UserID uuid.UUID `json:"user_id" query:"user_id" format:"uuid" validate:"required,uuid"`
} //	@name	UnbanUserRequest
