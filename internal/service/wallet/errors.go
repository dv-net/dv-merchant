package wallet

import "errors"

var ErrServiceWalletNotFound = errors.New("wallets not found")
var ErrAddressHasNoTransactions = errors.New("address has no transactions")
