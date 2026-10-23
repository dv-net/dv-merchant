package kraken

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
)

var (
	ErrInvalidNonce     = errors.New("kraken: invalid nonce")
	ErrUnknownAsset     = errors.New("kraken: unknown asset")
	ErrUnknownAssetPair = errors.New("kraken: unknown asset pair")
	ErrUnknownOrder     = errors.New("kraken: unknown order")
	ErrTooManyAddresses = errors.New("kraken: deposit address limit reached")
	ErrMaxFeeExceeded   = errors.New("kraken: withdrawal fee is above max_fee")
	ErrUnknownReference = errors.New("kraken: unknown withdrawal reference id")
)

// Kraken errors look like "<severity><category>:<message>[:<extra>]",
// e.g. "EGeneral:Invalid arguments:volume", so they are matched by prefix.
var sentinelByPrefix = []struct {
	prefix string
	err    error
}{
	{"EAPI:Invalid key", exchangeclient.ErrInvalidAPICredentials},
	{"EAPI:Invalid signature", exchangeclient.ErrInvalidAPICredentials},
	{"EAPI:Invalid nonce", ErrInvalidNonce},
	{"EQuery:Unknown asset pair", ErrUnknownAssetPair},
	{"EQuery:Unknown asset", ErrUnknownAsset},
	{"EOrder:Unknown order", ErrUnknownOrder},
	{"EGeneral:Permission denied", exchangeclient.ErrIncorrectAPIPermissions},
	{"EGeneral:Temporary lockout", exchangeclient.ErrRateLimited},
	{"EAPI:Rate limit exceeded", exchangeclient.ErrRateLimited},
	{"EOrder:Rate limit exceeded", exchangeclient.ErrRateLimited},
	{"EGeneral:Too many requests", exchangeclient.ErrRateLimited},
	{"EService:Throttled", exchangeclient.ErrRateLimited},
	{"EOrder:Insufficient funds", exchangeclient.ErrInsufficientBalance},
	{"EFunding:Insufficient funds", exchangeclient.ErrInsufficientBalance},
	{"EOrder:Orders minimum not met", exchangeclient.ErrMinOrderValue},
	{"EOrder:Cost minimum not met", exchangeclient.ErrMinOrderValue},
	{"EService:Market in cancel_only mode", exchangeclient.ErrSymbolTradingHalted},
	{"EService:Market in post_only mode", exchangeclient.ErrSymbolTradingHalted},
	{"EService:Market in limit_only mode", exchangeclient.ErrSymbolTradingHalted},
	{"EFunding:Unknown withdraw key", exchangeclient.ErrWithdrawalAddressNotWhitelisted},
	{"EFunding:Invalid amount", exchangeclient.ErrMinWithdrawalBalance},
	{"EFunding:Too many addresses", ErrTooManyAddresses},
	{"EFunding:Max fee exceeded", ErrMaxFeeExceeded},
	{"EFunding:Unknown reference id", ErrUnknownReference},
	{"EFunding:Invalid address", exchangeclient.ErrInvalidAddress},
}

func errorFromResponse(errs []string) error {
	msg := strings.Join(errs, "; ")
	for _, e := range errs {
		for _, s := range sentinelByPrefix {
			if strings.HasPrefix(e, s.prefix) {
				return fmt.Errorf("kraken error: %s: %w", msg, s.err)
			}
		}
	}
	return fmt.Errorf("kraken error: %s", msg)
}

func errorFromStatus(status int, body string) error {
	if status == http.StatusTooManyRequests {
		return fmt.Errorf("kraken error: status %d, body: %s: %w", status, sanitizeBody(body), exchangeclient.ErrRateLimited)
	}
	return fmt.Errorf("kraken error: status %d, body: %s", status, sanitizeBody(body))
}

func sanitizeHeaders(headers http.Header) map[string]string {
	sanitized := make(map[string]string, len(headers))
	for k, v := range headers {
		lk := strings.ToLower(k)
		if strings.Contains(lk, "key") || strings.Contains(lk, "sign") {
			sanitized[k] = "***REDACTED***"
		} else {
			sanitized[k] = strings.Join(v, ",")
		}
	}
	return sanitized
}

// sanitizeBody truncates long payloads for logging.
func sanitizeBody(body string) string {
	if len(body) == 0 {
		return "(empty)"
	}
	if len(body) > 500 {
		return body[:500] + "... (truncated)"
	}
	return body
}
