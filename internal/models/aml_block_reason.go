package models

import "github.com/shopspring/decimal"

// AmlBlockReason is one "reject" risk rule that fired on a deposit. It is a snapshot of the
// rule at the time of the check, the user's rules may be changed later.
type AmlBlockReason struct {
	// RiskType is the rule's risk_type: a signal category, TOTAL_RISK_SCORE, SUM_OF_SIGNALS or RISK_LEVEL
	RiskType string `json:"risk_type"`
	// Threshold is the rule's threshold, for RISK_LEVEL it is the level rank (none=0 ... critical=4)
	Threshold decimal.Decimal `json:"threshold"`
	// Value is the checked value which reached the threshold, for RISK_LEVEL it is the level rank
	Value decimal.Decimal `json:"value"`
} //	@name	AmlBlockReason

// AmlBlockReasons is the blocked_transactions.block_reasons column.
type AmlBlockReasons []AmlBlockReason //	@name	AmlBlockReasons
