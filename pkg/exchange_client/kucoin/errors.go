package kucoin

import (
	"fmt"
	"strings"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
)

// wrapKuCoinError wraps KuCoin errors with centralized errors when applicable
func wrapKuCoinError(msg string, code int) error {
	msgLower := strings.ToLower(msg)

	// Check for locked from withdrawal
	if strings.Contains(msgLower, "locked from withdrawal") {
		return fmt.Errorf("kucoin error: %s (%d): %w", msg, code, exchangeclient.ErrWithdrawalBalanceLocked)
	}
	if strings.Contains(msgLower, "funds were still awaiting confirmation") {
		return fmt.Errorf("kucoin errors: %s (%d): %w", msg, code, exchangeclient.ErrWithdrawalBalanceLocked)
	}

	return fmt.Errorf("kucoin error: %s (%d)", msg, code)
}
