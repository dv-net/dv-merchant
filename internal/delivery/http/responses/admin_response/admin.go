package admin_response

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

type AddUserRoleResponse struct {
	UserID    uuid.UUID         `json:"user_id" validate:"required" format:"uuid"`
	UserRoles []models.UserRole `json:"user_roles" validate:"required" enums:"root,user"`
} // @name AddUserRoleResponse

type RemoveUserRoleResponse struct {
	UserRoles []models.UserRole `json:"user_roles" format:"enum" enums:"root,user"`
} // @name RemoveUserRoleResponse

type UnbanUserResponse struct {
	UserID uuid.UUID `json:"user_id" validate:"required,uuid" format:"uuid"`
	Banned bool      `json:"banned"`
} // @name UnbanUserResponse

type GetUsersResponse struct {
	Email     string    `json:"email" format:"email"`
	CreatedAt time.Time `json:"created_at" format:"date-time"`
	UserID    uuid.UUID `json:"user_id" format:"uuid"`
	Roles     []string  `json:"roles" enums:"root,user"`
	Banned    bool      `json:"banned"`
} // @name GetUsersResponse

type BanUserResponse struct {
	UserID uuid.UUID `json:"user_id" format:"uuid"`
	Banned bool      `json:"banned"`
} // @name BanUserResponse

type InviteUserResponse struct {
	UserID uuid.UUID `json:"user_id" format:"uuid"`
	Email  string    `json:"email" format:"email"`
	Token  uuid.UUID `json:"token" format:"uuid"`
} // @name InviteUserResponse
