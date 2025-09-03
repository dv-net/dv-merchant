package user

import (
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/processing"

	"github.com/shopspring/decimal"
)

type OwnerData struct {
	IsAuthorized bool            `json:"is_authorized"`
	Balance      decimal.Decimal `json:"balance"`
	OwnerID      string          `json:"owner_id"`
	Telegram     *string         `json:"telegram,omitempty"`
}

type ResetPasswordDto struct {
	Code            int
	Email           string
	NewPassword     string
	ConfirmPassword string
}

type ChangeEmailConfirmationDto struct {
	Code                 string
	NewEmail             string
	NewEmailConfirmation string
}

type RegisterUserDTO struct {
	User      *models.User                  `json:"user"`
	OwnerInfo *processing.RegisterOwnerInfo `json:"owner_info"`
}
