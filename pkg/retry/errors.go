package retry

import "errors"

var (
	ErrRetry = errors.New("retry")
	ErrExit  = errors.New("exit")
)
