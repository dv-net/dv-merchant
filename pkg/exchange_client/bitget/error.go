//nolint:tagliatelle
package bitget

import (
	"errors"
	"strings"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
)

type ResponseError struct {
	Code        int64  `json:"code,string"`
	Msg         string `json:"msg"`
	RequestTime int64  `json:"requestTime"`
	Data        any    `json:"data,omitempty"`
}

func (e ResponseError) Error() string {
	return e.Msg
}

// Unwrap allows ErrorResponse to wrap centralized errors
func (e ResponseError) Unwrap() error {
	msgLower := strings.ToLower(e.Msg)

	// Check for temporarily frozen errors
	if strings.Contains(msgLower, "temporarily frozen") {
		return exchangeclient.ErrWithdrawalBalanceLocked
	}

	return nil
}

var (
	ErrParameterOrderID       = errors.New("parameter orderid error")
	ErrParameterClientOrderID = errors.New("parameter clientoid error")
)
