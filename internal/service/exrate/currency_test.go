package exrate_test

import (
	"testing"

	"github.com/dv-net/dv-merchant/internal/service/exrate"

	"github.com/stretchr/testify/assert"
)

func TestCurrencyPair(t *testing.T) {
	testData := []struct {
		name       string
		currencies []exrate.Currency
		pairs      []exrate.CurrencyPair
	}{
		{
			name: "single non-stable currency",
			currencies: []exrate.Currency{
				{Code: "BTC", IsStable: false},
			},
			pairs: []exrate.CurrencyPair{},
		},
		{
			name: "two non-stable currencies",
			currencies: []exrate.Currency{
				{Code: "BTC", IsStable: false},
				{Code: "ETH", IsStable: false},
			},
			pairs: []exrate.CurrencyPair{},
		},
		{
			name: "one stable and one non-stable",
			currencies: []exrate.Currency{
				{Code: "USD", IsStable: true},
				{Code: "BTC", IsStable: false},
			},
			pairs: []exrate.CurrencyPair{
				{From: "USD", To: "BTC"},
			},
		},
		{
			name: "multiple currencies with stables",
			currencies: []exrate.Currency{
				{Code: "BTC", IsStable: false},
				{Code: "ETH", IsStable: false},
				{Code: "USD", IsStable: true},
				{Code: "USDT", IsStable: true},
			},
			pairs: []exrate.CurrencyPair{
				{From: "BTC", To: "USD"},
				{From: "BTC", To: "USDT"},
				{From: "ETH", To: "USD"},
				{From: "ETH", To: "USDT"},
				{From: "USD", To: "USDT"},
			},
		},
		{
			name: "only stablecoins",
			currencies: []exrate.Currency{
				{Code: "USD", IsStable: true},
				{Code: "USDT", IsStable: true},
				{Code: "EUR", IsStable: true},
			},
			pairs: []exrate.CurrencyPair{
				{From: "USD", To: "USDT"},
				{From: "USD", To: "EUR"},
				{From: "USDT", To: "EUR"},
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			cl := exrate.CurrencyList(test.currencies)
			pairs := []exrate.CurrencyPair{}
			next := cl.Iter()
			for pair, ok := next(); ok; pair, ok = next() {
				pairs = append(pairs, pair)
			}
			assert.ElementsMatch(t, test.pairs, pairs,
				"unexpected currency pairs for test case: %s", test.name)
		})
	}
}
