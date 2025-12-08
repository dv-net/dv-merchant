package aml_responses

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AmlHistoryResponse struct {
	ID             uuid.UUID             `json:"id"`
	UserID         uuid.UUID             `json:"user_id"`
	ServiceID      uuid.UUID             `json:"service_id"`
	ServiceSlug    models.AMLSlug        `json:"service_slug"`
	ExternalID     string                `json:"external_id"`
	Status         models.AMLCheckStatus `json:"status"`
	Score          decimal.Decimal       `json:"score"`
	RiskLevel      *models.AmlRiskLevel  `json:"risk_level"`
	CreatedAt      *time.Time            `json:"created_at"`
	UpdatedAt      *time.Time            `json:"updated_at"`
	RequestHistory []CheckHistory        `json:"request_history"`
} //	@name	AmlHistoryResponse

type CheckHistory struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	AmlCheckID      uuid.UUID  `db:"aml_check_id" json:"aml_check_id"`
	RequestPayload  string     `db:"request_payload" json:"request_payload"`
	ServiceResponse string     `db:"service_response" json:"service_response"`
	ErrorMsg        *string    `db:"error_msg" json:"error_msg"`
	AttemptNumber   int32      `db:"attempt_number" json:"attempt_number"`
	CreatedAt       *time.Time `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
} //	@name	CheckHistory
