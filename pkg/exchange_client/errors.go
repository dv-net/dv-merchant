package exchangeclient

import "errors"

var (
	ErrInvalidIPAddress                = errors.New("incorrect IP/API permissions")
	ErrInvalidAPICredentials           = errors.New("invalid API credentials")
	ErrIncorrectAPIPermissions         = errors.New("incorrect API permissions")
	ErrSoftLockByUserSecurityAction    = errors.New("temporary disabled due to user security action") // password reset/change, etc.
	ErrWithdrawalBalanceLocked         = errors.New("withdrawal balance locked")
	ErrMinWithdrawalBalance            = errors.New("withdrawal threshold not met")
	ErrWithdrawalAddressNotWhitelisted = errors.New("withdrawal address not whitelisted")
	ErrWithdrawalPending               = errors.New("withdrawal pending")
	ErrInvalidAddress                  = errors.New("invalid address")
	ErrRateLimited                     = errors.New("rate limited")
)
