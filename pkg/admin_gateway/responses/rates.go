package responses

import (
	"time"

	"github.com/shopspring/decimal"
)

type RatesResponse struct {
	Source string          `json:"source"`
	From   string          `json:"from"`
	To     string          `json:"to"`
	Value  decimal.Decimal `json:"value"`
	Date   time.Time       `json:"date"`
}
