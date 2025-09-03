package bitok

import (
	"time"

	"github.com/shopspring/decimal"
)

// CheckTransferResponse represents the response from the /manual-checks/check-transfer/ endpoint.
type CheckTransferResponse struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	CheckType   string    `json:"check_type"`
	CheckStatus string    `json:"check_status"`
	CheckedAt   time.Time `json:"checked_at"`
	Transfer    struct {
		Network       string          `json:"network"`
		TokenID       string          `json:"token_id"`
		TokenSymbol   string          `json:"token_symbol"`
		TxStatus      string          `json:"tx_status"`
		TxHash        string          `json:"tx_hash"`
		OccurredAt    time.Time       `json:"occurred_at"`
		InputAddress  string          `json:"input_address"`
		OutputAddress string          `json:"output_address"`
		Direction     string          `json:"direction"`
		Amount        decimal.Decimal `json:"amount"`
		ValueInFiat   decimal.Decimal `json:"value_in_fiat"`
	} `json:"transfer"`
	Address struct {
		Network string `json:"network"`
		Address string `json:"address"`
	} `json:"address"`
	RiskLevel    string          `json:"risk_level"`
	RiskScore    decimal.Decimal `json:"risk_score"`
	FiatCurrency string          `json:"fiat_currency"`
}
