package exchangeclient

import "errors"

var (
	ErrInvalidIPAddress                = errors.New("incorrect IP/API permissions")
	ErrInvalidAPICredentials           = errors.New("invalid API credentials")
	ErrIncorrectAPIPermissions         = errors.New("incorrect API permissions")
	ErrSoftLockByUserSecurityAction    = errors.New("temporary disabled due to user security action") // password reset/change, etc.
	ErrWithdrawalBalanceLocked         = errors.New("withdrawal balance locked")
	ErrMinWithdrawalBalance            = errors.New("withdrawal threshold not met")
	ErrInsufficientBalance             = errors.New("insufficient balance")
	ErrSymbolTradingHalted             = errors.New("symbol trading is halted")
	ErrWithdrawalAddressNotWhitelisted = errors.New("withdrawal address not whitelisted")
	ErrWithdrawalPending               = errors.New("withdrawal pending")
	ErrInvalidAddress                  = errors.New("invalid address")
	ErrRateLimited                     = errors.New("rate limited")
	ErrMinOrderValue                   = errors.New("order value below minimum")
	ErrSkipOrder                       = errors.New("skip order") // custom error for all cases when we need to skip order
)
