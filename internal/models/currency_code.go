package models

const (
	CurrencyCodeUSD   string = "USD"
	CurrencyCodeUSDT  string = "USDT"
	CurrencyCodeUSDC  string = "USDC"
	CurrencyCodeDAI   string = "DAI"
	CurrencyCodeUSDD  string = "USDD"
	CurrencyCodeUSDE  string = "USDE"
	CurrencyCodeUSD1  string = "USD1"
	CurrencyCodePYUSD string = "PYUSD"
)

func CurrencyCodeStableList() []string {
	return []string{
		CurrencyCodeUSD,
		CurrencyCodeUSDT,
		CurrencyCodeUSDC,
		CurrencyCodeDAI,
		CurrencyCodeUSDD,
		CurrencyCodeUSDE,
		CurrencyCodeUSD1,
		CurrencyCodePYUSD,
	}
}

func CurrencyCodeStableSet() map[string]struct{} {
	return map[string]struct{}{
		CurrencyCodeUSD:   {},
		CurrencyCodeUSDT:  {},
		CurrencyCodeUSDC:  {},
		CurrencyCodeDAI:   {},
		CurrencyCodeUSDD:  {},
		CurrencyCodeUSDE:  {},
		CurrencyCodeUSD1:  {},
		CurrencyCodePYUSD: {},
	}
}
