package mexc

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
)

type ErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

var ErrDepositAddressExists = errors.New("mexc: chain does not support multiple deposit addresses")

func errorFromResponse(errRes *ErrorResponse) error {
	if sentinel := sentinelByCode(errRes.Code); sentinel != nil {
		return fmt.Errorf("mexc error: %s (%d): %w", errRes.Msg, errRes.Code, sentinel)
	}

	return fmt.Errorf("mexc error: %s (%d)", errRes.Msg, errRes.Code)
}

func errorFromStatus(status int, body string) error {
	if sentinel := sentinelByStatus(status); sentinel != nil {
		return fmt.Errorf("mexc error: status %d, body: %s: %w", status, sanitizeBody(body), sentinel)
	}

	return fmt.Errorf("mexc error: status %d, body: %s", status, sanitizeBody(body))
}

func sentinelByStatus(status int) error {
	switch status {
	case http.StatusTooManyRequests, http.StatusForbidden:
		return exchangeclient.ErrRateLimited
	default:
		return nil
	}
}

func sentinelByCode(code int) error {
	switch code {
	case 152073:
		return ErrDepositAddressExists
	case 602, 700002, 10072:
		return exchangeclient.ErrInvalidAPICredentials
	case 700006:
		return exchangeclient.ErrInvalidIPAddress
	case 401, 700007:
		return exchangeclient.ErrIncorrectAPIPermissions
	case 429, 403:
		return exchangeclient.ErrRateLimited
	case 10101, 30005:
		return exchangeclient.ErrInsufficientBalance
	case 30002:
		return exchangeclient.ErrMinOrderValue
	case 30000, 30016, 30018, 30019:
		return exchangeclient.ErrSymbolTradingHalted
	case 10212:
		return exchangeclient.ErrWithdrawalAddressNotWhitelisted
	case 10265, 60005:
		return exchangeclient.ErrSoftLockByUserSecurityAction
	default:
		return nil
	}
}

func sanitizeHeaders(headers http.Header) map[string]string {
	sanitized := make(map[string]string)
	for k, v := range headers {
		if strings.Contains(strings.ToLower(k), "key") ||
			strings.Contains(strings.ToLower(k), "sign") ||
			strings.Contains(strings.ToLower(k), "signature") {
			sanitized[k] = "***REDACTED***"
		} else {
			sanitized[k] = strings.Join(v, ",")
		}
	}
	return sanitized
}

// sanitizeBody truncates long request payloads for logging
func sanitizeBody(body string) string {
	if len(body) == 0 {
		return "(empty)"
	}
	if len(body) > 500 {
		return body[:500] + "... (truncated)"
	}
	return body
}
