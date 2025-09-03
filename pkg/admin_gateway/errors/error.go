package admin_errors

import (
	"errors"
	"fmt"
)

var ErrPrepareRequest = errors.New("prepare request")

type RequestFailedError struct {
	StatusCode int
	Body       []byte
	RequestURL string
}

func (re *RequestFailedError) Error() string {
	return fmt.Sprintf("request to %s failed with code %d and content: %s", re.RequestURL, re.StatusCode, re.Body)
}
