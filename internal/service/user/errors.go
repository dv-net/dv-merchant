package user

import "errors"

var (
	ErrOwnerIDIsNotSet             = errors.New("owner ID is not set")
	ErrClientIDNotFound            = errors.New("client ID not found")
	ErrAdminSecretNotFound         = errors.New("dvadmin secret not found")
	ErrTwoFactorAuthIsNotConfirmed = errors.New("two-factor auth data has not been confirmed")
	ErrInvalidOTP                  = errors.New("invalid otp")
	ErrDvTokenNotSet               = errors.New("dv token not set")
)
