package withdrawal_requests

import (
	"errors"

	"github.com/shopspring/decimal"
)

type CreateProcessingWithdrawRequest struct {
	Amount     decimal.Decimal `json:"amount" validate:"required"`
	AddressTo  string          `json:"address_to" validate:"required,min=16,max=255"`
	CurrencyID string          `json:"currency_id" validate:"required"`
	RequestID  *string         `json:"request_id" validate:"required,min=1,max=255"`
} //	@name	CreateProcessingWithdrawRequest

func (req *CreateProcessingWithdrawRequest) Validate() error {
	if !req.Amount.GreaterThan(decimal.Zero) {
		return errors.New("amount must be greater than zero")
	}

	return nil
}

type CreateProcessingWithdrawInternalRequest struct {
	Amount     decimal.Decimal `json:"amount" validate:"required"`
	AddressTo  string          `json:"address_to" validate:"required,min=16,max=255"`
	CurrencyID string          `json:"currency_id" validate:"required"`
	RequestID  *string         `json:"request_id" validate:"required"`
	TOTP       string          `json:"totp" validate:"required,min=6,max=6"`
} //	@name	CreateProcessingWithdrawalInternalRequest

func (req *CreateProcessingWithdrawInternalRequest) Validate() error {
	if !req.Amount.GreaterThan(decimal.Zero) {
		return errors.New("amount must be greater than zero")
	}

	if len(req.TOTP) != 6 {
		return errors.New("TOTP must be exactly 6 characters long")
	}

	if req.RequestID != nil && *req.RequestID == "" {
		return errors.New("request ID cannot be empty if provided")
	}

	return nil
}
