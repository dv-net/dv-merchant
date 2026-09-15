package store

import "errors"

var (
	ErrUserHasNoAccess     = errors.New("user has no access for current store")
	ErrStoreSecretNotFound = errors.New("store secret not found")
	ErrInvalidOTP          = errors.New("invalid OTP")
)
