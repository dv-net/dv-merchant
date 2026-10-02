package constants

const (
	AmlRiskTypeTotalScore   = "TOTAL_RISK_SCORE"
	AmlRiskTypeSumOfSignals = "SUM_OF_SIGNALS"
	// AmlRiskTypeRiskLevel compares the provider's risk level by rank (see models.AmlRiskLevel.Rank).
	AmlRiskTypeRiskLevel = "RISK_LEVEL"
)

const (
	AmlRiskRuleActionReject        = "reject"
	AmlRiskRuleActionAcceptAndFlag = "accept_and_flag"
)

const (
	AmlRiskRuleDefaultThreshold = 75
	AmlRiskRuleDefaultAction    = AmlRiskRuleActionReject
	// AmlRiskRuleDefaultLevelThreshold is the default RISK_LEVEL threshold: medium and above.
	AmlRiskRuleDefaultLevelThreshold = 2
)
