package processing

import (
	"errors"
	"strconv"
	"strings"

	"connectrpc.com/connect"
)

func ErrorRPCCode(err error) (int, bool) {
	if connectErr := new(connect.Error); errors.As(err, &connectErr) {
		for key, values := range connectErr.Meta() {
			if strings.ToLower(key) != "rpc-code" || len(values) != 1 {
				continue
			}

			if value, err := strconv.ParseInt(values[0], 10, 64); err == nil {
				return int(value), true
			}
		}
	}

	return 0, false
}
