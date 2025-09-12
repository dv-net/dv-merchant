package exchange_withdrawal

import "errors"

var (
	ErrThresholdNotMet                = errors.New("withdrawal threshold not met")
	ErrWithdrawalPending              = errors.New("withdrawal pending")
	ErrInsufficientBalance            = errors.New("insufficient balance")
	ErrWithdrawalBalanceLocked        = errors.New("withdrawal balance locked")
	ErrSoftLockByUserSecurityAction   = errors.New("temporary locked because of user security action")
	ErrWithdrawalAddessNotWhitelisted = errors.New("withdrawal address not whitelisted")
	ErrInvalidAddress                 = errors.New("invalid address")
)
