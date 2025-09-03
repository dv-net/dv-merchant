package errors

import (
	"github.com/dv-net/dv-merchant/internal/tools/apierror"
)

var ErrNoMatchesFound = apierror.Error{
	Message: "no matches found",
}
