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
	switch errRes.Code {
	case 152073:
		return ErrDepositAddressExists
	case 602, 700002, 10072:
		return exchangeclient.ErrInvalidAPICredentials
	case 700006:
		return exchangeclient.ErrInvalidIPAddress
	case 401, 700007:
		return exchangeclient.ErrIncorrectAPIPermissions
	default:
		return fmt.Errorf("mexc error: %s (%d)", errRes.Msg, errRes.Code)
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
