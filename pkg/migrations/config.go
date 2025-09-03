package migrations

import (
	"errors"
	"fmt"
)

type Config struct {
	DBDriver            DBDriver
	DSN                 string
	DisableConfirmation bool
}

func (c Config) Validate() error {
	var errs []error

	if !c.DBDriver.Valid() {
		errs = append(errs, fmt.Errorf("invalid db driver: %s", c.DBDriver))
	}

	if len(c.DSN) == 0 {
		errs = append(errs, fmt.Errorf("dsn is required"))
	}

	return errors.Join(errs...)
}
