package dto

import (
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/google/uuid"
)

type AddWalletAddressInPoolDTO struct {
	UserID   uuid.UUID        `json:"userId"`
	Currency *models.Currency `json:"currency"`
}
