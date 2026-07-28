package aml_test

import (
	"testing"

	"github.com/dv-net/dv-merchant/internal/constants"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/aml"
	externalaml "github.com/dv-net/dv-merchant/pkg/aml"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func rule(riskType string, threshold int64, action string, enabled bool) *models.UserAmlRiskRule {
	return &models.UserAmlRiskRule{
		RiskType:  riskType,
		Enabled:   enabled,
		Threshold: decimal.NewFromInt(threshold),
		Action:    action,
	}
}

func signal(category string, weight int64) externalaml.SignalContribution {
	return externalaml.SignalContribution{Category: category, Weight: decimal.NewFromInt(weight)}
}

func TestEvaluateRiskRules(t *testing.T) {
	t.Run("no rules means nothing is blocked or flagged", func(t *testing.T) {
		blocked, flagged := aml.EvaluateRiskRules(decimal.NewFromInt(90), nil, nil)
		require.False(t, blocked)
		require.False(t, flagged)
	})

	t.Run("disabled rule is ignored", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule(constants.AmlRiskTypeTotalScore, 50, constants.AmlRiskRuleActionReject, false),
		}
		blocked, flagged := aml.EvaluateRiskRules(decimal.NewFromInt(90), nil, rules)
		require.False(t, blocked)
		require.False(t, flagged)
	})

	t.Run("total score at or above threshold blocks", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule(constants.AmlRiskTypeTotalScore, 50, constants.AmlRiskRuleActionReject, true),
		}
		blocked, flagged := aml.EvaluateRiskRules(decimal.NewFromInt(50), nil, rules)
		require.True(t, blocked)
		require.True(t, flagged)
	})

	t.Run("total score below threshold does not fire", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule(constants.AmlRiskTypeTotalScore, 50, constants.AmlRiskRuleActionReject, true),
		}
		blocked, flagged := aml.EvaluateRiskRules(decimal.NewFromInt(49), nil, rules)
		require.False(t, blocked)
		require.False(t, flagged)
	})

	t.Run("accept_and_flag flags without blocking", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule(constants.AmlRiskTypeTotalScore, 50, constants.AmlRiskRuleActionAcceptAndFlag, true),
		}
		blocked, flagged := aml.EvaluateRiskRules(decimal.NewFromInt(60), nil, rules)
		require.False(t, blocked)
		require.True(t, flagged)
	})

	t.Run("category rule uses the matching signal weight", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule("SANCTIONS", 30, constants.AmlRiskRuleActionReject, true),
		}
		signals := []externalaml.SignalContribution{
			signal("SANCTIONS", 30),
			signal("GAMBLING", 100),
		}
		blocked, flagged := aml.EvaluateRiskRules(decimal.Zero, signals, rules)
		require.True(t, blocked)
		require.True(t, flagged)
	})

	t.Run("multiple contributions to the same category are summed", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule("SANCTIONS", 30, constants.AmlRiskRuleActionReject, true),
		}
		signals := []externalaml.SignalContribution{
			signal("SANCTIONS", 15),
			signal("SANCTIONS", 15),
		}
		blocked, _ := aml.EvaluateRiskRules(decimal.Zero, signals, rules)
		require.True(t, blocked)
	})

	t.Run("sum of signals only counts categories with their own enabled rule", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule("SANCTIONS", 30, constants.AmlRiskRuleActionReject, true),
			rule("GAMBLING", 20, constants.AmlRiskRuleActionReject, true),
			rule(constants.AmlRiskTypeSumOfSignals, 40, constants.AmlRiskRuleActionReject, true),
		}
		signals := []externalaml.SignalContribution{
			signal("SANCTIONS", 28),
			signal("GAMBLING", 19),
			signal("DARKNET", 100), // no rule for this category, must not count toward the sum
		}
		blocked, flagged := aml.EvaluateRiskRules(decimal.Zero, signals, rules)
		require.True(t, blocked, "SANCTIONS(28)+GAMBLING(19)=47 should trigger SUM_OF_SIGNALS>=40")
		require.True(t, flagged)
	})

	t.Run("sum of signals excludes disabled category rules", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule("SANCTIONS", 30, constants.AmlRiskRuleActionReject, true),
			rule("GAMBLING", 20, constants.AmlRiskRuleActionReject, false), // disabled, must not count
			rule(constants.AmlRiskTypeSumOfSignals, 40, constants.AmlRiskRuleActionReject, true),
		}
		signals := []externalaml.SignalContribution{
			signal("SANCTIONS", 28),
			signal("GAMBLING", 19),
		}
		blocked, flagged := aml.EvaluateRiskRules(decimal.Zero, signals, rules)
		require.False(t, blocked, "GAMBLING is excluded from the sum because its rule is disabled: 28 < 40")
		require.False(t, flagged)
	})

	t.Run("reject rule takes precedence over accept_and_flag for blocked, both still flag", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule("SANCTIONS", 30, constants.AmlRiskRuleActionAcceptAndFlag, true),
			rule("GAMBLING", 20, constants.AmlRiskRuleActionReject, true),
		}
		signals := []externalaml.SignalContribution{
			signal("SANCTIONS", 30),
			signal("GAMBLING", 20),
		}
		blocked, flagged := aml.EvaluateRiskRules(decimal.Zero, signals, rules)
		require.True(t, blocked)
		require.True(t, flagged)
	})
}
