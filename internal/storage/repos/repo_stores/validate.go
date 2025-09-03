package repo_stores

import (
	"errors"
)

func (s CreateParams) Validate() error {
	if s.Name == "" {
		return errors.New("name is required and cannot be empty")
	}

	if s.CurrencyID == "" {
		return errors.New("CurrencyID is required and cannot be empty")
	}

	if s.RateSource == "" {
		return errors.New("RateSource is required and cannot be empty")
	}

	return nil
}

func (s UpdateParams) Validate() error {
	if s.Name == "" {
		return errors.New("name is required and cannot be empty")
	}

	if s.CurrencyID == "" {
		return errors.New("CurrencyID is required and cannot be empty")
	}

	if s.RateSource == "" {
		return errors.New("RateSource is required and cannot be empty")
	}

	return nil
}
