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

// Rank returns the ordinal of the risk level (none=0 .. critical=4) used by RISK_LEVEL rules.
// Undefined and unknown levels have no rank.
func (l AmlRiskLevel) Rank() (int64, bool) {
	switch l {
	case AmlRiskLevelNone:
		return 0, true
	case AmlRiskLevelLow:
		return 1, true
	case AmlRiskLevelMedium:
		return 2, true
	case AmlRiskLevelHigh:
		return 3, true
	case AmlRiskLevelCritical:
		return 4, true
	default:
		return 0, false
	}
}
