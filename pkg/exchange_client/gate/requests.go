package gateio

type GetCurrencyChainsSupportedRequest struct {
	Currency string `json:"currency"`
}

type GetDepositAddressRequest struct {
	Currency string `json:"currency"`
}

type GetTickersInfoRequest struct {
	CurrencyPair string `json:"currency_pair,omitempty"`
}

type GetSpotAccountBalancesRequest struct {
	Currency string `json:"currency,omitempty"`
}

type GetWithdrawalHistoryRequest struct {
	Currency        string `json:"currency,omitempty" url:"currency,omitempty"`
	WithdrawalID    string `json:"withdrawal_id,omitempty" url:"withdrawal_id,omitempty"`
	AssetClass      string `json:"asset_class,omitempty" url:"asset_class,omitempty"`
	WithdrawOrderID string `json:"withdraw_order_id,omitempty" url:"withdraw_order_id,omitempty"`
	From            string `json:"from,omitempty" url:"from,omitempty"`
	To              string `json:"to,omitempty" url:"to,omitempty"`
	Limit           string `json:"limit,omitempty" url:"limit,omitempty"`
	Offset          string `json:"offset,omitempty" url:"offset,omitempty"`
}

type GetWithdrawalRulesRequest struct {
	Currency string `json:"currency,omitempty" url:"currency,omitempty"`
}

// buy means quote currency, BTC_USDT means USDT
// sell means base currency，BTC_USDT means BTC
type CreateSpotOrderRequest struct {
	CurrencyPair string `json:"currency_pair"`
	Type         string `json:"type"`   // "limit", "market", etc.
	Side         string `json:"side"`   // "buy" or "sell"
	Amount       string `json:"amount"` // Amount to buy/sell
	TimeInForce  string `json:"time_in_force"`
}

type CreateWithdrawalRequest struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Address  string `json:"address"`
	Chain    string `json:"chain"`
}

type GetWalletSavedAddressesRequest struct {
	Currency string `json:"currency" url:"currency"`
	Chain    string `json:"chain,omitempty" url:"chain,omitempty"`
}
