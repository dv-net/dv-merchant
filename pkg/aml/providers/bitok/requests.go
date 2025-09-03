package bitok

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

// CheckTransferRequest represents a request to the /manual-checks/check-transfer/ endpoint.
type CheckTransferRequest struct {
	Network       string    `json:"network"`
	TokenID       string    `json:"token_id"`
	TxHash        string    `json:"tx_hash"`
	OutputAddress string    `json:"output_address"`
	Direction     string    `json:"direction"`
	RiskModel     RiskModel `json:"risk_model,omitempty"`
}

// MarshalJSON implements json.Marshaler for CheckTransferRequest.
func (r *CheckTransferRequest) MarshalJSON() ([]byte, error) {
	type Alias CheckTransferRequest
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}

// RequestModel defines a request model that supports JSON serialization.
type RequestModel interface {
	json.Marshaler
}

// TransactionRequest represents a request to check a transaction for AML risks.
type TransactionRequest struct {
	TransactionID string          `json:"transaction_id"`
	Amount        decimal.Decimal `json:"amount"`
	RiskType      RiskModel       `json:"risk_type"`
}

// MarshalJSON implements json.Marshaler for TransactionRequest.
func (r *TransactionRequest) MarshalJSON() ([]byte, error) {
	type Alias TransactionRequest
	return json.Marshal(&struct {
		*Alias
		RiskType string `json:"risk_type"`
	}{
		Alias:    (*Alias)(r),
		RiskType: string(r.RiskType),
	})
}

// SenderRequest represents a request to check a sender entity for AML risks.
type SenderRequest struct {
	SenderID string    `json:"sender_id"`
	RiskType RiskModel `json:"risk_type"`
}

// MarshalJSON implements json.Marshaler for SenderRequest.
func (r *SenderRequest) MarshalJSON() ([]byte, error) {
	type Alias SenderRequest
	return json.Marshal(&struct {
		*Alias
		RiskType string `json:"risk_type"`
	}{
		Alias:    (*Alias)(r),
		RiskType: string(r.RiskType),
	})
}

// RecipientRequest represents a request to check a recipient entity for AML risks.
type RecipientRequest struct {
	RecipientID string    `json:"recipient_id"`
	RiskType    RiskModel `json:"risk_type"`
}

// MarshalJSON implements json.Marshaler for RecipientRequest.
func (r *RecipientRequest) MarshalJSON() ([]byte, error) {
	type Alias RecipientRequest
	return json.Marshal(&struct {
		*Alias
		RiskType string `json:"risk_type"`
	}{
		Alias:    (*Alias)(r),
		RiskType: string(r.RiskType),
	})
}
