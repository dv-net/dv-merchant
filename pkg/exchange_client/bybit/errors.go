//nolint:tagliatelle
package bybit

import (
	"errors"
	"fmt"
)

type APIError struct {
	Code    int64  `json:"retCode"`
	Message string `json:"retMsg"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("<APIError> code=%d, msg=%s", e.Code, e.Message)
}

func IsAPIError(e error) bool {
	var APIError *APIError
	ok := errors.As(e, &APIError)
	return ok
}
