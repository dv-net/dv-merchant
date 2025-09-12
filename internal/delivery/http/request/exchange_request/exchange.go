package exchange_request

type UpdateExchangePairsRequest struct {
	Pairs []ExchangePair `json:"pairs" validate:"required"`
} // @name ExchangeUpdatePairsRequest

type ExchangePair struct {
	BaseSymbol  string `json:"base_symbol" validate:"required"`
	QuoteSymbol string `json:"quote_symbol" validate:"required"`
	Symbol      string `json:"symbol" validate:"required"`
	Type        string `json:"type" validate:"required"`
} // @name ExchangePair

type GetDepositAddressRequest struct {
	Currency string `json:"currency" query:"currency" validate:"required"`
} // @name ExchangeGetDepositAddressRequest

type GetExchangeRulesRequest struct {
	Currency string `json:"currency"`
	Chain    string `json:"chain"`
} // @name ExchangeGetExchangeRulesRequest

type CreateWithdrawalSettingRequest struct {
	Address    string `json:"address" validate:"required"`
	MinAmount  string `json:"min_amount" validate:"required"`
	CurrencyID string `json:"currency_id" validate:"required"`
	Chain      string `json:"chain" validate:"required"`
} // @name ExchangeCreateWithdrawalRequest

type UpdateWithdrawalSetting struct {
	Enabled bool `json:"enabled"`
} // @name UpdateWithdrawalSetting

type GetWithdrawalsRequest struct {
	CurrencyID *string `json:"currency_id,omitempty" query:"currency_id" validate:"omitempty"`
	DateFrom   *string `json:"date_from" query:"date_from" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	DateTo     *string `json:"date_to" query:"date_to" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00,gtfield=DateFrom"`
	Slug       *string `json:"slug,omitempty" query:"slug" validate:"omitempty,alphanum"`
	Page       *uint32 `json:"page,omitempty" query:"page" validate:"omitempty,min=1"`
	PageSize   *uint32 `json:"page_size,omitempty" query:"page_size" validate:"omitempty,min=1,max=100"`
} // @name ExchangeGetWithdrawalsRequest

type GetWithdrawalsExportedRequest struct {
	CurrencyID *string `json:"currency_id,omitempty" query:"currency_id"`
	DateFrom   *string `json:"date_from" query:"date_from"`
	DateTo     *string `json:"date_to" query:"date_to"`
	Slug       *string `json:"slug,omitempty" query:"slug"`
	Format     string  `json:"format" validate:"required,oneof=csv xlsx" enums:"csv,xlsx"`
} // @name ExchangeGetWithdrawalsExportedRequest

type GetExchangeOrdersHistoryRequest struct {
	DateFrom *string `json:"date_from,omitempty" query:"date_from"`
	DateTo   *string `json:"date_to,omitempty" query:"date_to"`
	Page     *uint32 `json:"page" validate:"omitempty,numeric,gte=1"`
	PageSize *uint32 `json:"page_size" validate:"omitempty,min=1,max=100"`
	Slug     *string `json:"slug,omitempty" query:"slug"`
} // @name GetExchangeOrdersHistoryRequest

type ToggleExchangeWithdrawalsRequest struct {
	NewState string `json:"new_state" validate:"required,oneof=enabled disabled" enums:"enabled,disabled"`
} // @name ExchangeToggleWithdrawalsRequest

type ToggleExchangeSwapsRequest struct {
	NewState string `json:"new_state" validate:"required,oneof=enabled disabled" enums:"enabled,disabled"`
} // @name ExchangeToggleSwapsRequest

type GetExchangeOrdersHistoryExportedRequest struct {
	DateFrom *string `json:"date_from,omitempty" query:"date_from"`
	DateTo   *string `json:"date_to,omitempty" query:"date_to"`
	Page     *uint32 `json:"page" validate:"omitempty,numeric,gte=1"`
	PageSize *uint32 `json:"page_size" validate:"omitempty,min=1,max=100"`
	Format   string  `json:"format" validate:"required,oneof=csv xlsx" enums:"csv,xlsx"`
	Slug     *string `json:"slug,omitempty" query:"slug"`
} // @name GetExchangeOrdersHistoryExportedRequest

type TestConnectionRequest struct {
	Slug        string `json:"slug" validate:"required"`
	Credentials struct {
		Key        string `json:"key" validate:"required"`
		Secret     string `json:"secret" validate:"required"`
		Passphrase string `json:"passphrase"`
	} `json:"credentials" validate:"required"`
} // @name ExchangeTestConnectionRequest
