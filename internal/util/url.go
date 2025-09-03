package util

import (
	"net/url"

	"github.com/shopspring/decimal"
)

type TopUpParams struct {
	Amount   decimal.Decimal
	Currency *string
}

func EnrichTopUpURLByParams(formURL *url.URL, params TopUpParams) {
	query := url.Values{}
	if params.Amount.IsPositive() {
		query.Add("amount", params.Amount.String())
	}
	if params.Currency != nil {
		query.Add("currency", *params.Currency)
	}
	formURL.RawQuery = query.Encode()
}
