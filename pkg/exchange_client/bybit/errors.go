//nolint:tagliatelle
package bybit

import (
	"errors"
	"fmt"
	"strings"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
)

// Bybit API error codes
const (
	ErrCodeInsufficientBalance     = 110004 // Insufficient wallet balance
	ErrCodeInsufficientAvailable   = 110007 // Insufficient available balance
	ErrCodeInsufficientWithdrawBal = 110012 // Insufficient withdrawable balance
)

type APIError struct {
	Code    int64  `json:"retCode"`
	Message string `json:"retMsg"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("<APIError> code=%d, msg=%s", e.Code, e.Message)
}

// Unwrap allows APIError to wrap centralized errors
func (e APIError) Unwrap() error {
	// Check for insufficient balance error codes
	if e.Code == ErrCodeInsufficientBalance ||
		e.Code == ErrCodeInsufficientAvailable ||
		e.Code == ErrCodeInsufficientWithdrawBal {
		return exchangeclient.ErrWithdrawalBalanceLocked
	}

	// Check for insufficient balance in message
	msgLower := strings.ToLower(e.Message)
	if strings.Contains(msgLower, "insufficient") || strings.Contains(msgLower, "balance") {
		return exchangeclient.ErrWithdrawalBalanceLocked
	}

	return nil
}

func IsAPIError(e error) bool {
	var APIError *APIError
	ok := errors.As(e, &APIError)
	return ok
}
