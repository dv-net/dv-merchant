package exrate_response

type ExchangeRateResponse struct {
	Source string `json:"source"`
	From   string `json:"from"`
	To     string `json:"to"`
	Value  string `json:"value"`
} //	@name	ExchangeRateResponse
