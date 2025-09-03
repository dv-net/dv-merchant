package withdrawal_requests

import (
	"github.com/google/uuid"
)

type WithdrawalToProcessingRequest struct {
	WalletAddressID uuid.UUID `json:"wallet_address_id" validate:"required"`
	CurrencyID      string    `json:"currency_id" validate:"required"`
} // @name WithdrawalToProcessingRequest

type MultipleWithdrawalToProcessingRequest struct {
	WalletAddressIDs []uuid.UUID `json:"wallet_address_ids"` //nolint:tagliatelle
	Exclude          []uuid.UUID `json:"exclude"`
	CurrencyID       string      `json:"currency_id" validate:"required"`
} // @name MultipleWithdrawalToProcessingRequest
