package requests

type GetCurrencyList struct{}

type GetCurrency struct {
	Currency string `json:"-"`
	Chain    string `json:"chain" url:"chain"`
}
