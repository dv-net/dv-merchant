package errors //nolint:revive

import (
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
)

var (
	ErrUserBanned = apierror.Error{
		Message: "user banned",
	}
	ErrInvalidCredentials = apierror.Error{
		Message: "invalid credentials",
	}
)
