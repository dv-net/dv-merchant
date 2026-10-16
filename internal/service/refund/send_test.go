package refund

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
)

func TestIsAmbiguousProcessingError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"error before the call", errors.New("fetch blocked transaction: no rows"), false},
		{"business error of processing", connect.NewError(connect.CodeInternal, errors.New("refund already exists")), false},
		{"invalid argument", connect.NewError(connect.CodeInvalidArgument, errors.New("invalid totp")), false},
		{"processing unavailable", connect.NewError(connect.CodeUnavailable, errors.New("connection refused")), true},
		{"deadline exceeded", connect.NewError(connect.CodeDeadlineExceeded, errors.New("timeout")), true},
		{"unknown", connect.NewError(connect.CodeUnknown, errors.New("eof")), true},
		{"wrapped context deadline", fmt.Errorf("create refund in processing: %w", context.DeadlineExceeded), true},
		{"wrapped unavailable", fmt.Errorf("send: %w", connect.NewError(connect.CodeUnavailable, errors.New("x"))), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isAmbiguousProcessingError(tt.err))
		})
	}
}
