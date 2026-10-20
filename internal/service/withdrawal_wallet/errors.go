package withdrawal_wallet

import "errors"

var ErrDepositAddressNotAllowed = errors.New("deposit address cannot be used as withdrawal address")
