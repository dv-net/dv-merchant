//nolint:tagliatelle
package requests

type GetAllSupportedCurrenciesRequest struct {
	Timestamp int64 `json:"ts,omitempty" url:"ts,omitempty"`
}

type GetCurrencyReferenceRequest struct {
	Currency string `json:"currency" url:"currency,omitempty"`
}

type GetMarketSymbolsRequest struct {
	Symbols string `json:"symbols,omitempty" url:"symbols,omitempty"`
}
