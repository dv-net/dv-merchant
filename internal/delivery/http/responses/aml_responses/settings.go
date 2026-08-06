package aml_responses

import (
	"github.com/dv-net/dv-merchant/internal/constants"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/pkg/aml"
	"github.com/shopspring/decimal"
)

type AmlSettingsResponse struct {
	Enabled      bool            `json:"enabled"`
	ProviderSlug *models.AMLSlug `json:"provider_slug"`
} //	@name	AmlSettingsResponse

func NewAmlSettingsResponse(s *models.UserAmlSetting) AmlSettingsResponse {
	return AmlSettingsResponse{
		Enabled:      s.Enabled,
		ProviderSlug: s.ProviderSlug,
	}
}

type RiskRuleResponse struct {
	RiskType  string          `json:"risk_type"`
	Enabled   bool            `json:"enabled"`
	Threshold decimal.Decimal `json:"threshold"`
	Action    string          `json:"action"`
} //	@name	RiskRuleResponse

func NewRiskRuleResponse(r *models.UserAmlRiskRule) RiskRuleResponse {
	return RiskRuleResponse{
		RiskType:  r.RiskType,
		Enabled:   r.Enabled,
		Threshold: r.Threshold,
		Action:    r.Action,
	}
}

func MergeRiskRules(categories []aml.SignalCategory, rules []*models.UserAmlRiskRule) []RiskRuleResponse {
	byType := make(map[string]*models.UserAmlRiskRule, len(rules))
	for _, r := range rules {
		byType[r.RiskType] = r
	}

	riskTypes := make([]string, 0, len(categories)+2)
	riskTypes = append(riskTypes, constants.AmlRiskTypeTotalScore, constants.AmlRiskTypeSumOfSignals)
	for _, c := range categories {
		riskTypes = append(riskTypes, c.Category)
	}

	resp := make([]RiskRuleResponse, 0, len(riskTypes))
	for _, riskType := range riskTypes {
		if rule, ok := byType[riskType]; ok {
			resp = append(resp, NewRiskRuleResponse(rule))
			continue
		}
		resp = append(resp, RiskRuleResponse{
			RiskType:  riskType,
			Enabled:   false,
			Threshold: decimal.NewFromInt(constants.AmlRiskRuleDefaultThreshold),
			Action:    constants.AmlRiskRuleDefaultAction,
		})
	}
	return resp
}
