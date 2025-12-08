package admin_request

import (
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

type RemoveUserRoleRequest struct {
	UserID   uuid.UUID       `json:"user_id" query:"user_id" validate:"required" format:"uuid"`
	UserRole models.UserRole `json:"user_role" query:"user_role" validate:"required" enums:"root,user"`
} //	@name	RemoveUserRoleRequest
