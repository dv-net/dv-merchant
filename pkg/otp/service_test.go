package otp

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dv-net/dv-merchant/pkg/key_value"
	"github.com/stretchr/testify/require"
)

func newTestService(codes ...int) *Service {
	var i atomic.Int64
	gen := func() int {
		return codes[int(i.Add(1)-1)%len(codes)]
	}
	return New(&Config{TTL: time.Minute}, gen, key_value.NewInMemory())
}

func TestVerifyCode_SingleUse(t *testing.T) {
	ctx := context.Background()
	s := newTestService(123456)

	code, err := s.InitCode(ctx, "", "user@example.com")
	require.NoError(t, err)

	require.NoError(t, s.VerifyCode(ctx, code, "", "user@example.com"))
	require.ErrorIs(t, s.VerifyCode(ctx, code, "", "user@example.com"), ErrInvalidCode)
}

func TestVerifyCode_ScopedByPurpose(t *testing.T) {
	ctx := context.Background()
	s := newTestService(123456)

	code, err := s.InitCode(ctx, "", "a@example.com")
	require.NoError(t, err)

	require.ErrorIs(t, s.VerifyCode(ctx, code, "", "b@example.com"), ErrInvalidCode)
	require.NoError(t, s.VerifyCode(ctx, code, "", "a@example.com"))
}

func TestInitCode_ReplacesPreviousCode(t *testing.T) {
	ctx := context.Background()
	s := newTestService(111111, 222222)

	first, err := s.InitCode(ctx, "", "user@example.com")
	require.NoError(t, err)
	second, err := s.InitCode(ctx, "", "user@example.com")
	require.NoError(t, err)

	require.ErrorIs(t, s.VerifyCode(ctx, first, "", "user@example.com"), ErrInvalidCode)
	require.NoError(t, s.VerifyCode(ctx, second, "", "user@example.com"))
}

func TestVerifyCode_LockoutAfterMaxAttempts(t *testing.T) {
	ctx := context.Background()
	s := newTestService(123456)

	code, err := s.InitCode(ctx, "", "user@example.com")
	require.NoError(t, err)

	for range DefaultMaxAttempts {
		require.ErrorIs(t, s.VerifyCode(ctx, 999999, "", "user@example.com"), ErrInvalidCode)
	}

	// Even the right code is rejected once the limit is reached, and the code is burned.
	require.ErrorIs(t, s.VerifyCode(ctx, code, "", "user@example.com"), ErrTooManyAttempts)
}

func TestInitCode_DoesNotResetAttempts(t *testing.T) {
	ctx := context.Background()
	s := newTestService(111111, 222222)

	_, err := s.InitCode(ctx, "", "user@example.com")
	require.NoError(t, err)
	for range DefaultMaxAttempts {
		require.Error(t, s.VerifyCode(ctx, 999999, "", "user@example.com"))
	}

	// Requesting a fresh code must not give the attacker a fresh set of guesses.
	fresh, err := s.InitCode(ctx, "", "user@example.com")
	require.NoError(t, err)
	require.ErrorIs(t, s.VerifyCode(ctx, fresh, "", "user@example.com"), ErrTooManyAttempts)
}

func TestVerifyCode_SuccessResetsAttempts(t *testing.T) {
	ctx := context.Background()
	s := newTestService(111111, 222222)

	first, err := s.InitCode(ctx, "", "user@example.com")
	require.NoError(t, err)
	for range DefaultMaxAttempts - 1 {
		require.Error(t, s.VerifyCode(ctx, 999999, "", "user@example.com"))
	}
	require.NoError(t, s.VerifyCode(ctx, first, "", "user@example.com"))

	second, err := s.InitCode(ctx, "", "user@example.com")
	require.NoError(t, err)
	require.Error(t, s.VerifyCode(ctx, 999999, "", "user@example.com"))
	require.NoError(t, s.VerifyCode(ctx, second, "", "user@example.com"))
}

func TestVerifyCode_ConcurrentGuessesAreBounded(t *testing.T) {
	ctx := context.Background()
	s := newTestService(123456)

	_, err := s.InitCode(ctx, "", "user@example.com")
	require.NoError(t, err)

	var checked atomic.Int64
	var wg sync.WaitGroup
	for range 200 {
		wg.Go(func() {
			if err := s.VerifyCode(ctx, 999999, "", "user@example.com"); errors.Is(err, ErrInvalidCode) {
				checked.Add(1)
			}
		})
	}
	wg.Wait()

	// Only MaxAttempts guesses may actually be compared against the code.
	require.Equal(t, int64(DefaultMaxAttempts), checked.Load())
}

func TestStringCode_LockoutAfterMaxAttempts(t *testing.T) {
	ctx := context.Background()
	s := newTestService(0)
	gen := func() (string, error) { return "ABC123", nil }

	code, err := s.InitStringCode(ctx, "", "wallet:1", gen)
	require.NoError(t, err)

	for range DefaultMaxAttempts {
		require.ErrorIs(t, s.VerifyStringCode(ctx, "ZZZ999", "", "wallet:1"), ErrInvalidCode)
	}
	require.ErrorIs(t, s.VerifyStringCode(ctx, code, "", "wallet:1"), ErrTooManyAttempts)
}
