package aml

import (
	"fmt"

	"github.com/dv-net/dv-merchant/internal/constants"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/pkg/aml"
	"github.com/shopspring/decimal"
)

func convertAmlRiskLevelToModel(riskLevel aml.CheckRiskLevel) (*models.AmlRiskLevel, error) {
	var resolvedLevel models.AmlRiskLevel
	switch riskLevel {
	case aml.CheckRiskLevelLow:
		resolvedLevel = models.AmlRiskLevelLow
	case aml.CheckRiskLevelMedium:
		resolvedLevel = models.AmlRiskLevelMedium
	case aml.CheckRiskLevelHigh:
		resolvedLevel = models.AmlRiskLevelHigh
	case aml.CheckRiskLevelSevere:
		resolvedLevel = models.AmlRiskLevelCritical
	case aml.CheckRiskLevelUndefined:
		resolvedLevel = models.AmlRiskLevelUndefined
	case aml.CheckRiskLevelNone:
		resolvedLevel = models.AmlRiskLevelNone
	default:
		return nil, fmt.Errorf("unknown external risk level '%s'", string(riskLevel))
	}

	return &resolvedLevel, nil
}

func convertAmlStatusToModel(status aml.CheckStatus) models.AMLCheckStatus {
	switch status {
	case aml.CheckStatusSuccess:
		return models.AmlCheckStatusSuccess
	case aml.CheckStatusFailure:
		return models.AmlCheckStatusFailed
	default:
		return models.AmlCheckStatusPending
	}
}

// EvaluateRiskRules checks a deposit against the user's configured risk rules.
// score is the provider's aggregate score (used by AmlRiskTypeTotalScore rules);
// signals is the per-category breakdown (used by category rules and summed for
// AmlRiskTypeSumOfSignals, which only counts categories that have their own enabled
// rule). Disabled rules are ignored; rule order doesn't matter, since the sum is fully
// computed before any rule is evaluated against its threshold.
//
// Example: SANCTIONS{30,reject} and GAMBLING{20,reject} both pass individually
// (28 < 30, 19 < 20), but SUM_OF_SIGNALS{40,reject} still fires on 28+19=47, so
// EvaluateRiskRules returns (true, true).
//
// Returns blocked (a "reject" rule fired — send the AML-blocked webhook instead of the
// normal one) and flagged (any rule fired at all, reject or accept_and_flag — mark the
// address dirty regardless of which).
func EvaluateRiskRules(score decimal.Decimal, signals []aml.SignalContribution, rules []*models.UserAmlRiskRule) (blocked, flagged bool) {
	signalWeights := make(map[string]decimal.Decimal, len(signals))
	for _, s := range signals {
		signalWeights[s.Category] = signalWeights[s.Category].Add(s.Weight)
	}

	var categorySum decimal.Decimal
	for _, rule := range rules {
		if rule.Enabled && rule.RiskType != constants.AmlRiskTypeTotalScore && rule.RiskType != constants.AmlRiskTypeSumOfSignals {
			categorySum = categorySum.Add(signalWeights[rule.RiskType])
		}
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		var actual decimal.Decimal
		switch rule.RiskType {
		case constants.AmlRiskTypeTotalScore:
			actual = score
		case constants.AmlRiskTypeSumOfSignals:
			actual = categorySum
		default:
			actual = signalWeights[rule.RiskType]
		}
		if actual.GreaterThanOrEqual(rule.Threshold) {
			if rule.Action == constants.AmlRiskRuleActionReject {
				blocked = true
			} else {
				flagged = true
			}
		}
	}
	return blocked, blocked || flagged
}
