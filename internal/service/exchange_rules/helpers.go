package exchange_rules

import (
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_chains"

	"github.com/samber/lo"
)

func unwrapCurrencies(currencies []*repo_exchange_chains.GetEnabledCurrenciesRow) []string {
	return lo.Map(currencies, func(c *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) string {
		return c.ID.String
	})
}

type RuleType string

func (o RuleType) String() string { return string(o) }

const (
	WithdrawalRuleType RuleType = "withdrawal"
	SpotOrderRuleType  RuleType = "spot_order"
)

func formatKey(exchangeSlug models.ExchangeSlug, ruleType RuleType, currency string, userID string) string {
	return "exchange_rules" + "_" + ruleType.String() + ":" + exchangeSlug.String() + ":" + currency + ":" + userID
}
