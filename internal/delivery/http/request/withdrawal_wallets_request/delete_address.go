package withdrawal_wallets_request

import (
	"github.com/google/uuid"
)

type DeleteAddressRequest struct {
	TOTP string `json:"totp" validate:"required,len=6"`
} //	@name	DeleteWithdrawalAddressRequest

type BatchDeleteAddressRequest struct {
	TOTP       string      `json:"totp" validate:"required,len=6"`
	AddressIDs []uuid.UUID `json:"address_ids" validate:"required"` //nolint:tagliatelle
} //	@name	BatchDeleteWithdrawalAddressRequest
