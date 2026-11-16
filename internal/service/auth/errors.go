package auth

import "errors"

var (
	ErrTokenExpired   = errors.New("token expired")
	ErrWalletNotFound = errors.New("wallet not found")
	ErrEmailMismatch  = errors.New("email does not match wallet")
	ErrInvalidOTPCode = errors.New("invalid or expired code")

	ErrWalletEmailMissing      = errors.New("wallet has no email")
	ErrRefundCodeCooldown      = errors.New("refund confirmation code was sent recently")
	ErrRefundCodeAttemptsLimit = errors.New("too many refund confirmation code attempts")
)
