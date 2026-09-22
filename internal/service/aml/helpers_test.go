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

// flagRule builds an "accept_and_flag" rule. Its own risk_type is what ends up in
// matchedFlags — there's no separate canonical flag value to pick.
func flagRule(riskType string, threshold int64, enabled bool) *models.UserAmlRiskRule {
	return rule(riskType, threshold, constants.AmlRiskRuleActionAcceptAndFlag, enabled)
}

func signal(category string, weight int64) externalaml.SignalContribution {
	return externalaml.SignalContribution{Category: category, Weight: decimal.NewFromInt(weight)}
}

func TestEvaluateRiskRules(t *testing.T) {
	t.Run("no rules means nothing is blocked or flagged", func(t *testing.T) {
		blocked, flags := aml.EvaluateRiskRules(decimal.NewFromInt(90), nil, nil)
		require.False(t, blocked)
		require.Empty(t, flags)
	})

	t.Run("disabled rule is ignored", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule(constants.AmlRiskTypeTotalScore, 50, constants.AmlRiskRuleActionReject, false),
		}
		blocked, flags := aml.EvaluateRiskRules(decimal.NewFromInt(90), nil, rules)
		require.False(t, blocked)
		require.Empty(t, flags)
	})

	t.Run("total score at or above threshold blocks", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule(constants.AmlRiskTypeTotalScore, 50, constants.AmlRiskRuleActionReject, true),
		}
		blocked, flags := aml.EvaluateRiskRules(decimal.NewFromInt(50), nil, rules)
		require.True(t, blocked)
		require.Empty(t, flags)
	})

	t.Run("total score below threshold does not fire", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule(constants.AmlRiskTypeTotalScore, 50, constants.AmlRiskRuleActionReject, true),
		}
		blocked, flags := aml.EvaluateRiskRules(decimal.NewFromInt(49), nil, rules)
		require.False(t, blocked)
		require.Empty(t, flags)
	})

	t.Run("accept_and_flag flags without blocking, tagged with its own risk_type", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			flagRule(constants.AmlRiskTypeTotalScore, 50, true),
		}
		blocked, flags := aml.EvaluateRiskRules(decimal.NewFromInt(60), nil, rules)
		require.False(t, blocked)
		require.Equal(t, []string{constants.AmlRiskTypeTotalScore}, flags)
	})

	t.Run("accept_and_flag below its own threshold does not fire", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			flagRule("SANCTIONS", 30, true),
		}
		signals := []externalaml.SignalContribution{signal("SANCTIONS", 20)}
		blocked, flags := aml.EvaluateRiskRules(decimal.Zero, signals, rules)
		require.False(t, blocked)
		require.Empty(t, flags)
	})

	t.Run("category rule uses the matching signal weight", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule("SANCTIONS", 30, constants.AmlRiskRuleActionReject, true),
		}
		signals := []externalaml.SignalContribution{
			signal("SANCTIONS", 30),
			signal("GAMBLING", 100),
		}
		blocked, _ := aml.EvaluateRiskRules(decimal.Zero, signals, rules)
		require.True(t, blocked)
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
		blocked, _ := aml.EvaluateRiskRules(decimal.Zero, signals, rules)
		require.True(t, blocked, "SANCTIONS(28)+GAMBLING(19)=47 should trigger SUM_OF_SIGNALS>=40")
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
		blocked, _ := aml.EvaluateRiskRules(decimal.Zero, signals, rules)
		require.False(t, blocked, "GAMBLING is excluded from the sum because its rule is disabled: 28 < 40")
	})

	t.Run("sum of signals excludes accept_and_flag category rules", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			rule("SANCTIONS", 30, constants.AmlRiskRuleActionReject, true),
			flagRule("GAMBLING", 15, true), // accept_and_flag, must not count toward the sum
			rule(constants.AmlRiskTypeSumOfSignals, 40, constants.AmlRiskRuleActionReject, true),
		}
		signals := []externalaml.SignalContribution{
			signal("SANCTIONS", 28),
			signal("GAMBLING", 19),
		}
		blocked, flags := aml.EvaluateRiskRules(decimal.Zero, signals, rules)
		// Without the fix, categorySum would be 28+19=47 >= 40 and blocked would wrongly become true.
		require.False(t, blocked, "GAMBLING is excluded from the sum because its rule is accept_and_flag, not reject: 28 < 40")
		require.Equal(t, []string{"GAMBLING"}, flags, "the accept_and_flag rule still fires independently on its own category weight (19 >= 15)")
	})

	t.Run("reject and accept_and_flag rules act independently", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			flagRule("SANCTIONS", 30, true),
			rule("GAMBLING", 20, constants.AmlRiskRuleActionReject, true),
		}
		signals := []externalaml.SignalContribution{
			signal("SANCTIONS", 30),
			signal("GAMBLING", 20),
		}
		blocked, flags := aml.EvaluateRiskRules(decimal.Zero, signals, rules)
		require.True(t, blocked, "GAMBLING reject rule fires")
		require.Equal(t, []string{"SANCTIONS"}, flags, "SANCTIONS accept_and_flag rule fires independently, address is not blocked by it")
	})

	t.Run("multiple accept_and_flag rules produce multiple distinct flags", func(t *testing.T) {
		rules := []*models.UserAmlRiskRule{
			flagRule("exchange_sanctioned_eu", 10, true),
			flagRule("DARKNET_MARKETPLACE", 10, true),
		}
		signals := []externalaml.SignalContribution{
			signal("exchange_sanctioned_eu", 15),
			signal("DARKNET_MARKETPLACE", 15),
		}
		blocked, flags := aml.EvaluateRiskRules(decimal.Zero, signals, rules)
		require.False(t, blocked)
		require.ElementsMatch(t, []string{"exchange_sanctioned_eu", "DARKNET_MARKETPLACE"}, flags)
	})

	t.Run("a risk_type appearing more than once is only reported once", func(t *testing.T) {
		// Not reachable through the API today (user_id+provider_slug+risk_type is unique
		// at the DB level), but EvaluateRiskRules itself should stay safe either way.
		rules := []*models.UserAmlRiskRule{
			flagRule("SANCTIONS", 10, true),
			flagRule("SANCTIONS", 10, true),
		}
		signals := []externalaml.SignalContribution{signal("SANCTIONS", 15)}
		blocked, flags := aml.EvaluateRiskRules(decimal.Zero, signals, rules)
		require.False(t, blocked)
		require.Equal(t, []string{"SANCTIONS"}, flags)
	})
}
