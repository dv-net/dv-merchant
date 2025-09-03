package aml

import (
	"errors"
	"fmt"
)

var ErrPrepareRequest = errors.New("invalid reuq")

type RequestFailedError struct {
	StatusCode int
	Body       []byte
	RequestURL string
}

func (re *RequestFailedError) Error() string {
	return fmt.Sprintf("request to %s failed with code %d and content: %s", re.RequestURL, re.StatusCode, re.Body)
}
