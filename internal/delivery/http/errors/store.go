package errors

import (
	"errors"
)

var (
	ErrStoreNotFound      = errors.New("store not found")
	ErrStoreOwnerMismatch = errors.New("store owner mismatch")
)
