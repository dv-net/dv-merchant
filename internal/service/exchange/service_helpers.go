package exchange

import (
	"errors"
	"syscall"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
)

var (
	// businessErrors are expected business logic errors that should be ignored
	businessErrors = []error{
		// Balance and trading errors
		exchangeclient.ErrInsufficientBalance,
		exchangeclient.ErrSymbolTradingHalted,
		exchangeclient.ErrMinOrderValue,

		// Rate limiting
		exchangeclient.ErrRateLimited,

		// Credential errors - should not suspend transfers
		// Unusual case
		exchangeclient.ErrInvalidAPICredentials,
		exchangeclient.ErrInvalidIPAddress,
		exchangeclient.ErrIncorrectAPIPermissions,

		// Custom skip
		exchangeclient.ErrSkipOrder,
	}

	// networkErrors are transient network errors that should be ignored
	networkErrors = []error{
		// Connection errors
		syscall.ECONNRESET,   // Connection reset by peer
		syscall.ECONNREFUSED, // Connection refused
		syscall.ECONNABORTED, // Connection aborted
		syscall.ENETUNREACH,  // Network unreachable
		syscall.EHOSTUNREACH, // Host unreachable
		syscall.EHOSTDOWN,    // Host is down

		// I/O errors
		syscall.EPIPE, // Broken pipe
		syscall.EIO,   // I/O error

		// Timeout errors
		syscall.ETIMEDOUT, // Connection timed out
	}
)

// shouldErrorSuspendTransfers checks if the error should be silently ignored during order processing
func shouldErrorSuspendTransfers(err error) bool {
	if err == nil {
		return false
	}

	// Check business errors
	for _, ignorableErr := range businessErrors {
		if errors.Is(err, ignorableErr) {
			return false
		}
	}

	// Check network errors
	for _, ignorableErr := range networkErrors {
		if errors.Is(err, ignorableErr) {
			return false
		}
	}

	return true
}
