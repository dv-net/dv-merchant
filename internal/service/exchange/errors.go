package exchange

import "errors"

var (
	ErrExchangeNotFound        = errors.New("exchange not found")
	ErrUnsupportedExchangeType = errors.New("invalid exchange type")
	ErrInsufficientBalance     = errors.New("insufficient balance")
	ErrInvalidIPAddress        = errors.New("invalid IP address")
	ErrSymbolTradingHalted     = errors.New("symbol trading is halted")
	ErrSkipOrder               = errors.New("skip order")
)
