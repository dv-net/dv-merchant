package bitok

import "github.com/dv-net/dv-merchant/pkg/aml"

type CheckStatus string

const (
	CheckStatusChecked  CheckStatus = "checked"
	CheckStatusChecking CheckStatus = "checking"
	CheckStatusError    CheckStatus = "error"
)

func (s CheckStatus) ToAMLStatus() aml.CheckStatus {
	switch s {
	case CheckStatusChecked:
		return aml.CheckStatusSuccess
	case CheckStatusError:
		return aml.CheckStatusFailure
	default:
		return aml.CheckStatusNew
	}
}

type Direction string

const (
	DirectionIncoming Direction = "incoming"
	DirectionOutgoing Direction = "outgoing"
)

func (d Direction) String() string {
	return string(d)
}

func DirectionFromAML(direction aml.Direction) Direction {
	switch direction {
	case aml.DirectionIn:
		return DirectionIncoming
	default:
		return DirectionOutgoing
	}
}

type RiskModel string

const (
	RiskModelSenderEntity             RiskModel = "sender_entity"
	RiskModelRecipientEntity          RiskModel = "recipient_entity"
	RiskModelOriginOfFunds            RiskModel = "origin_of_funds"
	RiskModelDestinationOfFunds       RiskModel = "destination_of_funds"
	RiskModelSenderExposure           RiskModel = "sender_exposure"
	RiskModelRecipientExposure        RiskModel = "recipient_exposure"
	RiskModelAttemptSenderEntity      RiskModel = "attempt_sender_entity"
	RiskModelAttemptRecipientEntity   RiskModel = "attempt_recipient_entity"
	RiskModelAttemptSenderExposure    RiskModel = "attempt_sender_exposure"
	RiskModelAttemptRecipientExposure RiskModel = "attempt_recipient_exposure"
)

// signalCategories is sourced from GET /v1/basics/entity-categories/ (43 items).
var signalCategories = []aml.SignalCategory{
	{Category: "exchange_sanctioned_eu", Label: "Exchange, Sanctioned (EU)"},
	{Category: "exchange_sanctioned_us", Label: "Exchange, Sanctioned (US)"},
	{Category: "exchange_sanctioned_gb", Label: "Exchange, Sanctioned (GB)"},
	{Category: "atm", Label: "ATM"},
	{Category: "bridge", Label: "Bridge"},
	{Category: "child_abuse_material", Label: "Child Abuse Material"},
	{Category: "custodial_wallet", Label: "Custodial wallet"},
	{Category: "dust", Label: "Dust"},
	{Category: "dex", Label: "DEX"},
	{Category: "enforcement_action", Label: "Enforcement action"},
	{Category: "fraud_shop", Label: "Fraud shop"},
	{Category: "gambling", Label: "Gambling"},
	{Category: "high_risk_exchange", Label: "High-Risk Exchange"},
	{Category: "high_risk_jurisdiction", Label: "High-Risk Jurisdiction"},
	{Category: "iaas", Label: "IaaS"},
	{Category: "ico", Label: "ICO"},
	{Category: "illegal_service", Label: "Illegal Service"},
	{Category: "lending", Label: "Lending"},
	{Category: "mining", Label: "Mining"},
	{Category: "mining_pool", Label: "Mining Pool"},
	{Category: "mixer", Label: "Mixer"},
	{Category: "online_pharmacy", Label: "Online pharmacy"},
	{Category: "other", Label: "Other"},
	{Category: "payment_service_provider", Label: "Payment Service Provider"},
	{Category: "darknet_market", Label: "Darknet Market"},
	{Category: "exchange", Label: "Exchange"},
	{Category: "fee", Label: "Fee"},
	{Category: "nft_marketplace", Label: "NFT marketplace"},
	{Category: "sanctions", Label: "Sanctions"},
	{Category: "scam", Label: "Scam"},
	{Category: "seized_funds", Label: "Seized funds"},
	{Category: "smart_contract", Label: "Smart contract"},
	{Category: "stolen_funds", Label: "Stolen Funds"},
	{Category: "token_contract", Label: "Token contract"},
	{Category: "undefined", Label: "Undefined"},
	{Category: "unnamed_service", Label: "Unnamed service"},
	{Category: "marketplace", Label: "Marketplace"},
	{Category: "p2p_exchange", Label: "P2P exchange"},
	{Category: "privacy_protocol", Label: "Privacy protocol"},
	{Category: "ransomware", Label: "Ransomware"},
	{Category: "personal_wallet", Label: "Personal wallet"},
	{Category: "terrorist_financing", Label: "Terrorist Financing"},
	{Category: "unnamed_wallet", Label: "Unnamed wallet"},
}

var _ aml.SignalCategoryLister = (*Client)(nil)

func (c *Client) SignalCategories() []aml.SignalCategory {
	return signalCategories
}
