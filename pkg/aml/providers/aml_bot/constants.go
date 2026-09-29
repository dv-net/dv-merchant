package aml_bot

import "github.com/dv-net/dv-merchant/pkg/aml"

type CheckStatus string

const (
	CheckStatusSuccess CheckStatus = "success"
	CheckStatusPending CheckStatus = "pending"
	CheckStatusFailed  CheckStatus = "failed"
	CheckStatusError   CheckStatus = "error"
)

func (s CheckStatus) ToAMLStatus() aml.CheckStatus {
	switch s {
	case CheckStatusSuccess:
		return aml.CheckStatusSuccess
	case CheckStatusError, CheckStatusFailed:
		return aml.CheckStatusFailure
	default:
		return aml.CheckStatusNew
	}
}

type Direction string

const (
	DirectionDeposit    Direction = "deposit"
	DirectionWithdrawal Direction = "withdrawal"
)

func (d Direction) String() string {
	return string(d)
}

func DirectionFromAML(direction aml.Direction) Direction {
	switch direction {
	case aml.DirectionIn:
		return DirectionDeposit
	default:
		return DirectionWithdrawal
	}
}

// signalCategories is sourced from GET /signals/ (33 items), ordered by signalGroups: low, middle, high.
var signalCategories = []aml.SignalCategory{
	// low
	{Category: "merchant_services", Label: "Merchant services"},
	{Category: "exchange", Label: "Exchange"},
	{Category: "wallet", Label: "Wallet"},
	{Category: "miner", Label: "Miner"},
	{Category: "payment", Label: "Payment"},
	{Category: "marketplace", Label: "Marketplace"},
	{Category: "p2p_exchange", Label: "P2P exchange"},
	{Category: "seized_assets", Label: "Seized assets"},
	{Category: "other", Label: "Other"},
	{Category: "unnamed_entity_low_risk", Label: "Unnamed entity (low risk)"},
	// middle
	{Category: "infrastructure_as_a_service", Label: "Infrastructure as a service"},
	{Category: "decentralized_exchange_contract", Label: "Decentralized exchange contract"},
	{Category: "unnamed_service", Label: "Unnamed service"},
	{Category: "atm", Label: "ATM"},
	{Category: "risky_exchange", Label: "Risky exchange"},
	{Category: "p2p_exchange_mlrisk_high", Label: "P2P exchange, high ML risk"},
	{Category: "liquidity_pools", Label: "Liquidity pools"},
	{Category: "unnamed_entity_medium_risk", Label: "Unnamed entity (medium risk)"},
	// high
	{Category: "enforcement_action", Label: "Enforcement action"},
	{Category: "ransom", Label: "Ransom"},
	{Category: "dark_market", Label: "Dark market"},
	{Category: "dark_service", Label: "Dark service"},
	{Category: "illegal_service", Label: "Illegal service"},
	{Category: "mixer", Label: "Mixer"},
	{Category: "scam", Label: "Scam"},
	{Category: "gambling", Label: "Gambling"},
	{Category: "stolen_coins", Label: "Stolen coins"},
	{Category: "exchange_fraudulent", Label: "Fraudulent exchange"},
	{Category: "child_exploitation", Label: "Child exploitation"},
	{Category: "sanctions", Label: "Sanctions"},
	{Category: "terrorism_financing", Label: "Terrorism financing"},
	{Category: "malware", Label: "Malware"},
	{Category: "unnamed_entity_high_risk", Label: "Unnamed entity (high risk)"},
}

var _ aml.SignalCategoryLister = (*Client)(nil)

func (c *Client) SignalCategories() []aml.SignalCategory {
	return signalCategories
}
