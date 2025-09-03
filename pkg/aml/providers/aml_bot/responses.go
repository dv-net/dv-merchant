//nolint:tagliatelle
package aml_bot

import (
	"github.com/shopspring/decimal"
)

// ErrorResponse response with error from AMLBot
type ErrorResponse struct {
	Result      bool   `json:"result"`
	Description string `json:"description"`
}

// Response generic AMLBot response
type Response struct {
	Result               bool            `json:"result"`
	Description          string          `json:"description,omitempty"`
	Balance              decimal.Decimal `json:"balance"`
	AMLFlow              string          `json:"amlFlow"`
	TermsConfirmRequired bool            `json:"termsConfirmRequired"`
	ProMode              string          `json:"proMode"`
	Discount             decimal.Decimal `json:"discount"`
	DiscountDueTime      decimal.Decimal `json:"discountDueTime"`
	PromoFlow            bool            `json:"promoFlow"`
	Data                 *CheckData      `json:"data,omitempty"`
}

// CheckData AMLBot check/recheck data response
type CheckData struct {
	RiskScore             decimal.Decimal `json:"riskscore"`
	Signals               Signals         `json:"signals"`
	UpdatedAt             decimal.Decimal `json:"updated_at"`
	Address               string          `json:"address"`
	CreatedAt             decimal.Decimal `json:"created_at"`
	Amount                decimal.Decimal `json:"amount"`
	RiskyVolume           decimal.Decimal `json:"risky_volume"`
	Direction             string          `json:"direction"`
	Tx                    string          `json:"tx"`
	RiskyVolumeFiat       decimal.Decimal `json:"risky_volume_fiat"`
	Fiat                  decimal.Decimal `json:"fiat"`
	FiatCodeEffective     string          `json:"fiat_code_effective"`
	Counterparty          Counterparty    `json:"counterparty"`
	BlackListsConnections bool            `json:"blackListsConnections"`
	HasBlackListFlag      bool            `json:"hasBlackListFlag"`
	PdfReport             string          `json:"pdfReport"`
	Memo                  string          `json:"memo"`
	CustomerIsB2B         bool            `json:"customerIsB2B"`
	ConfirmedAt           decimal.Decimal `json:"confirmed_at"`
	UID                   string          `json:"uid"`
	Asset                 string          `json:"asset"`
	Network               string          `json:"network"`
	Status                string          `json:"status"`
	Timestamp             string          `json:"timestamp"`
	Flow                  string          `json:"flow"`
	Type                  int             `json:"_type"`
	Cost                  decimal.Decimal `json:"cost"`
}

// Signals AMLBot signals data
type Signals struct {
	Exchange                      decimal.Decimal `json:"exchange"`
	RiskyExchange                 decimal.Decimal `json:"risky_exchange"`
	P2PExchange                   decimal.Decimal `json:"p2p_exchange"`
	EnforcementAction             decimal.Decimal `json:"enforcement_action"`
	ATM                           decimal.Decimal `json:"atm"`
	ChildExploitation             decimal.Decimal `json:"child_exploitation"`
	DarkMarket                    decimal.Decimal `json:"dark_market"`
	DarkService                   decimal.Decimal `json:"dark_service"`
	ExchangeFraudulent            decimal.Decimal `json:"exchange_fraudulent"`
	Gambling                      decimal.Decimal `json:"gambling"`
	IllegalService                decimal.Decimal `json:"illegal_service"`
	LiquidityPools                decimal.Decimal `json:"liquidity_pools"`
	Marketplace                   decimal.Decimal `json:"marketplace"`
	Miner                         decimal.Decimal `json:"miner"`
	Mixer                         decimal.Decimal `json:"mixer"`
	Other                         decimal.Decimal `json:"other"`
	P2PExchangeMLRiskHigh         decimal.Decimal `json:"p2p_exchange_mlrisk_high"`
	Payment                       decimal.Decimal `json:"payment"`
	Ransom                        decimal.Decimal `json:"ransom"`
	Sanctions                     decimal.Decimal `json:"sanctions"`
	Scam                          decimal.Decimal `json:"scam"`
	SeizedAssets                  decimal.Decimal `json:"seized_assets"`
	StolenCoins                   decimal.Decimal `json:"stolen_coins"`
	TerrorismFinancing            decimal.Decimal `json:"terrorism_financing"`
	Wallet                        decimal.Decimal `json:"wallet"`
	InfrastructureAsAService      decimal.Decimal `json:"infrastructure_as_a_service"`
	DecentralizedExchangeContract decimal.Decimal `json:"decentralized_exchange_contract"`
	MerchantServices              decimal.Decimal `json:"merchant_services"`
	UnnamedService                decimal.Decimal `json:"unnamed_service"`
	Malware                       decimal.Decimal `json:"malware"`
}

// Counterparty AMLBot counterparty data
type Counterparty struct {
	ID                 interface{}         `json:"id"`
	ReceivedFiatAmount decimal.Decimal     `json:"received_fiat_amount"`
	SentFiatAmount     decimal.Decimal     `json:"sent_fiat_amount"`
	Signals            CounterpartySignals `json:"signals"`
}

// CounterpartySignals counterparty signals
type CounterpartySignals struct {
	In  Signals `json:"in"`
	Out Signals `json:"out"`
}
