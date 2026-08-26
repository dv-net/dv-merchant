package auth

import "errors"

var (
	ErrTokenExpired   = errors.New("token expired")
	ErrWalletNotFound = errors.New("wallet not found")
	ErrEmailMismatch  = errors.New("email does not match wallet")
	ErrInvalidOTPCode = errors.New("invalid or expired code")
)
