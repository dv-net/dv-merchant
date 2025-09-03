package models

type AmlRiskLevel string

const (
	AmlRiskLevelNone      = "none"
	AmlRiskLevelLow       = "low"
	AmlRiskLevelMedium    = "medium"
	AmlRiskLevelHigh      = "high"
	AmlRiskLevelCritical  = "critical"
	AmlRiskLevelUndefined = "undefined"
)
