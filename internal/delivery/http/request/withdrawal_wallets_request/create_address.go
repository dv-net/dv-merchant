package withdrawal_wallets_request

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
)

type CreateAddressRequest struct {
	Name    *string `json:"name" validate:"max=100,omitempty"`
	Address string  `json:"address" validate:"required,max=100"`
	TOTP    string  `json:"totp" validate:"required,len=6"`
} //	@name	CreateWithdrawalAddressRequest

type BatchCreateAddressRequest struct {
	Addresses string `json:"addresses" validate:"required"`
	TOTP      string `json:"totp" validate:"required,len=6"`
} //	@name	BatchCreateWithdrawalAddressRequest

type TransferRequest struct {
	Kinds    []models.TransferKind  `json:"kinds" validate:"omitempty,dive"`
	Stages   []models.TransferStage `json:"stages" query:"stages" validate:"required,dive"`
	DateFrom *time.Time             `json:"date_from" query:"date_from" validate:"omitempty" format:"date-time,omitempty"`
	Page     *uint32                `json:"page" query:"page"`
	PageSize *uint32                `json:"page_size" query:"page_size"`
} //	@name	WithdrawalWalletsTransferRequest
