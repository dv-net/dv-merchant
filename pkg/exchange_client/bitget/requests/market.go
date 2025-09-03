package requests

type (
	TickerInformationRequest struct {
		Symbol string `json:"symbol" url:"symbol,omitempty"`
	}

	CoinInformationRequest struct {
		Coin string `json:"coin" url:"coin,omitempty"`
	}
	SymbolInformationRequest struct {
		Symbol string `json:"symbol" url:"symbol,omitempty"`
	}
)
