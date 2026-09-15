package wallet

import "errors"

var (
	ErrServiceWalletNotFound    = errors.New("wallets not found")
	ErrAddressHasNoTransactions = errors.New("address has no transactions")
)
