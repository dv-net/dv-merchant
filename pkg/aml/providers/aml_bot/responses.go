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

// Signals AMLBot signals data: category -> share of funds (0..1)
type Signals map[string]decimal.Decimal

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
