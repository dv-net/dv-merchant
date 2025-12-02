package admin_request

import (
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

type InviteUserWithRoleRequest struct {
	Email    string          `json:"email" validate:"required,email" format:"email"`
	Role     models.UserRole `json:"role" validate:"required"`
	StoreIDs []uuid.UUID     `json:"store_ids" validate:"required"` //nolint:tagliatelle
	Mnemonic string          `json:"mnemonic" validate:"required,mnemonic"`
} //	@name	InviteUserWithRoleRequest
