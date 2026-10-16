package processing

import "errors"

var (
	ErrFundsWithdrawal            = errors.New("funds withdrawal error")
	ErrProcessingResourceExceeded = errors.New("processing resource exceeded")
	ErrServiceNotInitialized      = errors.New("service not initialized")
	ErrTransferNotFound           = errors.New("transfer not found in processing")
)
