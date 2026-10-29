package exchange

import (
	"errors"
)

var (
	ErrExchangeNotFound         = errors.New("exchange not found")
	ErrExchangeKeysNotSet       = errors.New("set exchange keys before network settings")
	ErrNetworkInterfaceNotFound = errors.New("network interface not found on the server")
	ErrNetworkSettingsCheck     = errors.New("exchange is unreachable with the new network settings")
)
