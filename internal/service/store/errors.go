package store

import "errors"

var ErrUserHasNoAccess = errors.New("user has no access for current store")
var ErrStoreSecretNotFound = errors.New("store secret not found")
var ErrInvalidOTP = errors.New("invalid OTP")
