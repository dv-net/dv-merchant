package withdraw

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

var (
	ErrStoreIsNotOwnedByUser                    = errors.New("store is not owned by user")
	ErrWalletIsNotOwnedByUser                   = errors.New("wallet is not owned by user")
	ErrFiatCurrencyIsNotSupported               = errors.New("fiat currencies is not supported")
	ErrTransfersDisabled                        = errors.New("transfers disabled by owner")
	ErrWithdrawalsFromProcessingDisabled        = errors.New("withdrawal from processing is disabled by owner")
	ErrWithdrawalNotFound                       = errors.New("withdrawal not found")
	ErrWithdrawalCannotBeDeleted                = errors.New("withdrawal cannot be deleted")
	ErrProcessingUnavailable                    = errors.New("processing is unavailable")
	ErrProcessingUninitialized                  = errors.New("processing is uninitialized")
	ErrProcessingWalletNotExists                = errors.New("processing wallet not exists")
	ErrTransferFromMultipleAddressNotSupported  = errors.New("multiple transfers accepted only for bitcoin-like blockchains")
	ErrWithdrawFromProcessingToHotNotAllowed    = errors.New("withdrawal to hot wallet is not allowed")
	ErrWithdrawFromProcessingDuplicateRequestID = errors.New("withdrawal with such request id already exists")
	ErrWithdrawalAddressListEmpty               = errors.New("withdrawal address list is empty")
	ErrWithdrawalAddressEmptyBalances           = errors.New("withdrawal addresses have empty balances")
	ErrProcessingExplorerUnavailable            = errors.New("explorer is unavailable")
)

type InvalidCurrencyForAddressError struct {
	Wallet     string
	Blockchain string
}

func (e *InvalidCurrencyForAddressError) Error() string {
	return fmt.Sprintf("invalid wallet '%s' for blockchain '%s'", e.Wallet, e.Blockchain)
}

func isIgnoredLogError(err error) bool {
	return errors.Is(err, ErrTransfersDisabled) ||
		errors.Is(err, pgx.ErrNoRows) ||
		errors.Is(err, ErrWithdrawalsFromProcessingDisabled) ||
		errors.Is(err, ErrWithdrawalAddressListEmpty)
}
