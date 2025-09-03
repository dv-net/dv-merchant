package key_value

import (
	"errors"
)

var (
	ErrEntryNotFound        = errors.New("not exists by key")
	ErrCounterLimitExceeded = errors.New("counter limit exceeded")
)
