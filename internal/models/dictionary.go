package models

type Dictionary struct {
	AvailableCurrencies []*CurrencyShort `json:"available_currencies"`
} //	@name	Dictionary

type ExchangeChainShort struct {
	CurrencyID  string `json:"currency_id"`
	Ticker      string `json:"ticker"`
	TickerLabel string `json:"ticker_label"`
	Chain       string `json:"chain"`
	ChainLabel  string `json:"chain_label"`
} //	@name	ExchangeChain
