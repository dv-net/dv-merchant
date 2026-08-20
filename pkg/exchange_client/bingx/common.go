package bingx

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
)

const successCode = 0

var ErrDepositAddressUnavailable = errors.New("bingx: deposit address query failed")

type ErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func errorFromResponse(errRes *ErrorResponse) error {
	if sentinel := sentinelByCode(errRes.Code, errRes.Msg); sentinel != nil {
		return fmt.Errorf("bingx error: %s (%d): %w", errRes.Msg, errRes.Code, sentinel)
	}

	return fmt.Errorf("bingx error: %s (%d)", errRes.Msg, errRes.Code)
}

func sentinelByCode(code int, msg string) error {
	switch code {
	case 100202:
		return exchangeclient.ErrInsufficientBalance
	case 100437:
		if strings.Contains(strings.ToLower(msg), "insufficient balance") {
			return exchangeclient.ErrInsufficientBalance
		}

		return ErrDepositAddressUnavailable
	case 100001, 100413:
		return exchangeclient.ErrInvalidAPICredentials
	case 100419:
		return exchangeclient.ErrInvalidIPAddress
	case 100403:
		return exchangeclient.ErrIncorrectAPIPermissions
	case 100410, 100440:
		return exchangeclient.ErrRateLimited
	case 100490:
		return exchangeclient.ErrSymbolTradingHalted
	case 100441:
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

func sanitizeBody(body string) string {
	if len(body) == 0 {
		return "(empty)"
	}
	if len(body) > 500 {
		return body[:500] + "... (truncated)"
	}
	return body
}
