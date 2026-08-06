package constants

const (
	AmlRiskTypeTotalScore   = "TOTAL_RISK_SCORE"
	AmlRiskTypeSumOfSignals = "SUM_OF_SIGNALS"
)

const (
	AmlRiskRuleActionReject        = "reject"
	AmlRiskRuleActionAcceptAndFlag = "accept_and_flag"
)

const (
	AmlRiskRuleDefaultThreshold = 75
	AmlRiskRuleDefaultAction    = AmlRiskRuleActionReject
)
