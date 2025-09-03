//nolint:tagliatelle
package htx

import (
	"encoding/json"
	"io"

	htxresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/responses"
)

type ResponseV1[T any] struct {
	Status    htxresponses.Status `json:"status,omitempty"`
	Code      int                 `json:"code,omitempty"`
	Message   string              `json:"message,omitempty"`
	Data      T                   `json:"data,omitempty"`
	Timestamp int64               `json:"timestamp,omitempty"`
	Full      int                 `json:"full,omitempty"`
	ErrCode   string              `json:"err-code,omitempty"`
	ErrMsg    string              `json:"err-msg,omitempty"`
}

type ResponseV2[T any] struct {
	Status    htxresponses.Status `json:"status,omitempty"`
	Code      int                 `json:"code,omitempty"`
	Message   string              `json:"message,omitempty"`
	Data      T                   `json:"data,omitempty"`
	Timestamp int64               `json:"timestamp,omitempty"`
	Full      int                 `json:"full,omitempty"`
	ErrCode   string              `json:"err-code,omitempty"`
	ErrMsg    string              `json:"err-msg,omitempty"`
}

type ResponseV1Error struct {
	Status     htxresponses.Status `json:"status,omitempty"`
	StatusCode int                 `json:"code,omitempty"`
	ErrCode    any                 `json:"err-code,omitempty"`
	ErrMsg     string              `json:"err-msg,omitempty"`
	Data       any                 `json:"data,omitempty"`
}

type ResponseV2Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func ParseResponseV1[T any](rc io.ReadCloser) (*ResponseV1[T], error) {
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	r := &ResponseV1[T]{}
	if err := json.Unmarshal(b, r); err != nil {
		return nil, err
	}
	return r, nil
}

func ParseResponseV1Error(rc io.ReadCloser) (*ResponseV1Error, error) {
	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	r := &ResponseV1Error{}
	if err := json.Unmarshal(b, r); err != nil {
		return nil, err
	}
	return r, nil
}
