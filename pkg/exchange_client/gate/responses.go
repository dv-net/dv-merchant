package gateio

type GetSpotCurrenciesResponse struct {
	Data []*CurrencyDetails `json:"data"`
}

type GetSpotCurrencyResponse struct {
	Data *CurrencyDetails `json:"data"`
}

type GetSpotSupportedCurrencyPairsResponse struct {
	Data []*CurrencyPair `json:"data"`
}

type GetSpotSupportedCurrencyPairResponse struct {
	Data *CurrencyPair `json:"data"`
}

type GetTickersInfoResponse struct {
	Data []*TickerInfo `json:"data"`
}

type GetAccountDetailResponse struct {
	Data *AccountDetail `json:"data"`
}

type GetCurrencySupportedChainResponse struct {
	Data []*CurrencyChain `json:"data"`
}

type GetDepositAddressResponse struct {
	Data *DepositAddress `json:"data"`
}

type GetSpotAccountBalancesResponse struct {
	Data []*SpotAccountBalance `json:"data"`
}

type GetWithdrawalHistoryResponse struct {
	Data []*WithdrawalHistory `json:"data"`
}

type GetWithdrawalRulesResponse struct {
	Data []*WithdrawalRule `json:"data"`
}

type CreateSpotOrderResponse struct {
	Data *SpotOrder `json:"data"`
}

type GetSpotOrderResponse struct {
	Data *SpotOrder `json:"data"`
}

type CreateWithdrawalResponse struct {
	Data *Withdrawal `json:"data"`
}

type GetWalletSavedAddressesResponse struct {
	Data []*SavedAddress `json:"data"`
}
