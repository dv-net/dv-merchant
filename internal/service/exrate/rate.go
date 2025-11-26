package exrate

import (
	"time"

	"github.com/shopspring/decimal"
)

type ExRate struct {
	Source     string     `json:"source"`
	From       string     `json:"from"`
	To         string     `json:"to"`
	Value      string     `json:"value"`
	ValueScale string     `json:"value_scale"`
	UpdatedAt  *time.Time `json:"updated_at" format:"date-time"`
} //	@name	ExchangeRate

type Rates struct {
	CurrencyIDs []string          `json:"currency_ids"` //nolint:tagliatelle
	Rate        []decimal.Decimal `json:"rate"`
} //	@name	ExchangeRates
