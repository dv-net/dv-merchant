package exchangeclient

import "errors"

var (
	ErrInvalidIPAddress             = errors.New("incorrect IP/API permissions")
	ErrInvalidAPICredentials        = errors.New("invalid API credentials")
	ErrIncorrectAPIPermissions      = errors.New("incorrect API permissions")
	ErrSoftLockByUserSecurityAction = errors.New("temporary disabled due to user security action") // password reset/change, etc.
)
