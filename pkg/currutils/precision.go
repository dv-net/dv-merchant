package currutils

import "github.com/shopspring/decimal"

func ValidateDecimalPrecision(amount decimal.Decimal, precision int16) bool {
	// Get the number of decimal places in the amount
	exponent := amount.Exponent()

	// If exponent is non-negative, there are no decimal places
	if exponent >= 0 {
		return true
	}

	// Count decimal places (exponent is negative)
	decimalPlaces := -exponent

	// Check if decimal places exceed the allowed precision
	return decimalPlaces <= int32(precision)
}
