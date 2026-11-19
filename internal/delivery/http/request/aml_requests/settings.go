package aml_requests

import (
	"github.com/shopspring/decimal"
)

type UpdateAmlSettingsRequest struct {
	Enabled           bool             `json:"enabled"`
	ProviderSlug      string           `json:"provider_slug"`
	MinCheckAmountUsd *decimal.Decimal `json:"min_check_amount_usd,omitempty" validate:"omitempty,decimal_gte=0"`
}

type RiskRuleRequest struct {
	RiskType  string          `json:"risk_type" validate:"required"`
	Enabled   bool            `json:"enabled"`
	Threshold decimal.Decimal `json:"threshold" validate:"required,gte=0,lte=100"`
	Action    string          `json:"action" validate:"required,oneof=reject accept_and_flag"`
} //	@name	RiskRuleRequest

type UpsertRiskRulesRequest struct {
	Rules []RiskRuleRequest `json:"rules" validate:"required,dive"`
} //	@name	UpsertRiskRulesRequest
