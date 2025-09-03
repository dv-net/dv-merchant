package models

const (
	CurrencyCodeUSD  string = "USD"
	CurrencyCodeUSDT string = "USDT"
	CurrencyCodeUSDC string = "USDC"
	CurrencyCodeDAI  string = "DAI"
)

func CurrencyCodeStableList() []string {
	return []string{
		CurrencyCodeUSD, CurrencyCodeUSDT, CurrencyCodeUSDC, CurrencyCodeDAI,
	}
}

func CurrencyCodeStableSet() map[string]struct{} {
	return map[string]struct{}{
		CurrencyCodeUSD:  {},
		CurrencyCodeUSDT: {},
		CurrencyCodeUSDC: {},
		CurrencyCodeDAI:  {},
	}
}
