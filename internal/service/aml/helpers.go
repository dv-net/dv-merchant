package aml

import (
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/pkg/aml"
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
