//nolint:tagliatelle
package bitget

import "errors"

type ErrorResponse struct {
	Code        int64  `json:"code,string"`
	Msg         string `json:"msg"`
	RequestTime int64  `json:"requestTime"`
	Data        any    `json:"data,omitempty"`
}

var (
	ErrParameterOrderID       = errors.New("parameter orderid error")
	ErrParameterClientOrderID = errors.New("parameter clientoid error")
)
