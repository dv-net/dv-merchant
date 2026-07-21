package aml

import (
	"fmt"

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

func subtractIgnoredSignals(score decimal.Decimal, signals []aml.SignalContribution, ignored []string) decimal.Decimal {
	if len(ignored) == 0 || len(signals) == 0 {
		return score
	}
	ignoredSet := make(map[string]struct{}, len(ignored))
	for _, c := range ignored {
		ignoredSet[c] = struct{}{}
	}
	for _, s := range signals {
		if _, ok := ignoredSet[s.Category]; ok {
			score = score.Sub(s.Weight)
		}
	}
	return decimal.Max(score, decimal.Zero)
}
