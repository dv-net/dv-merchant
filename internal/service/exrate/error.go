package exrate

import "fmt"

type ExchangeRateNotFoundError struct {
	Source, From, To string
}

func (e *ExchangeRateNotFoundError) Error() string {
	return fmt.Sprintf(
		"currency exchange rate from [%s] to [%s] not found in source [%s]",
		e.From, e.To, e.Source,
	)
}
