package aml

import "errors"

var ErrUnsupportedProvider = errors.New("unsupported or disabled provider")
var ErrUnsupportedCurrencies = errors.New("currency is not supported by provider")
var ErrInvalidAddress = errors.New("invalid address for blockchain")
