package models

type ExchangeRate struct {
	Source string `json:"source"`
	From   string `json:"from"`
	To     string `json:"to"`
	Value  string `json:"value"`
} // @name ExchangeRate
