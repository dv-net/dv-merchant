package aml

import "errors"

var (
	ErrUnsupportedProvider   = errors.New("unsupported or disabled provider")
	ErrUnsupportedCurrencies = errors.New("currency is not supported by provider")
	ErrInvalidAddress        = errors.New("invalid address for blockchain")
	ErrNoProviderAvailable   = errors.New("no aml provider available for this user and currency")
)
