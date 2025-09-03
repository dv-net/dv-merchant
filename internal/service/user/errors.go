package user

import "errors"

var ErrOwnerIDIsNotSet = errors.New("owner ID is not set")
var ErrClientIDNotFound = errors.New("client ID not found")
var ErrAdminSecretNotFound = errors.New("dvadmin secret not found")
var ErrTwoFactorAuthIsNotConfirmed = errors.New("two-factor auth data has not been confirmed")
var ErrInvalidOTP = errors.New("invalid otp")
var ErrDvTokenNotSet = errors.New("dv token not set")
