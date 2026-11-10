package withdrawal_wallet

import "errors"

var (
	ErrDepositAddressNotAllowed = errors.New("deposit address cannot be used as withdrawal address")
	ErrAddressesNotFound        = errors.New("some addresses not found in withdrawal wallet")
)
