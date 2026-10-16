package withdrawal_requests

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ManualWithdrawRequest struct {
	WalletAddressID uuid.UUID `json:"wallet_address_id" required:"true"`
	CurrencyID      string    `json:"currency_id" required:"true"`
} //	@name	ManualWithdrawRequest

type ManualMultipleWithdrawRequest struct {
	WithdrawalWalletID         uuid.UUID   `json:"withdrawal_wallet_id" required:"true"`
	WalletAddressIDs           []uuid.UUID `json:"wallet_address_ids"`           //nolint:tagliatelle
	ExcludedWalletAddressesIDs []uuid.UUID `json:"exclude_wallet_addresses_ids"` //nolint:tagliatelle
	CurrencyID                 string      `json:"currency_id" required:"true"`
} //	@name	ManualMultipleWithdrawRequest

type ExternalWithdrawRequest struct {
	WalletAddressID uuid.UUID       `json:"wallet_address_id" validate:"required"`
	AddressTo       string          `json:"address_to" validate:"required,min=16,max=255"`
	Amount          decimal.Decimal `json:"amount" validate:"required"`
	TOTP            string          `json:"totp" validate:"required,len=6"`
} //	@name	ExternalWithdrawRequest
