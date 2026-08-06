package coinkyt

import "github.com/dv-net/dv-merchant/pkg/aml"

var signalCategories = []aml.SignalCategory{
	{Category: "REWARDS_FEES", Label: "Rewards/Fees"},
	{Category: "UNKNOWN_OWNER", Label: "Unknown owner"},
	{Category: "ATM", Label: "ATM"},
	{Category: "DARKNET_MARKETPLACE", Label: "Darknet marketplace"},
	{Category: "DARKNET_SERVICE", Label: "Darknet service"},
	{Category: "ILLEGAL_SERVICE", Label: "Illegal service"},
	{Category: "SCAM", Label: "Scam"},
	{Category: "SCAM_CRYPTO_EXCHANGE", Label: "Scam crypto exchange"},
	{Category: "GAMBLING", Label: "Gambling"},
	{Category: "STOLEN_ASSETS", Label: "Stolen assets"},
	{Category: "MIXING_SERVICE", Label: "Mixing service"},
	{Category: "RANSOM", Label: "Ransom"},
	{Category: "EXCHANGE_LICENSED", Label: "Licensed exchange service"},
	{Category: "EXCHANGE_UNLICENSED", Label: "Unlicensed exchange service"},
	{Category: "P2P_EXCHANGE_UNLICENSED", Label: "P2P Exchange unlicensed"},
	{Category: "MINER", Label: "Miner"},
	{Category: "ONLINE_MARKETPLACE", Label: "Online marketplace"},
	{Category: "ONLINE_WALLET", Label: "Online wallet"},
	{Category: "P2P_EXCHANGE_LICENSED", Label: "P2P Exchange licensed"},
	{Category: "PAYMENT_SYSTEM", Label: "Payment system"},
	{Category: "OTHER", Label: "Other"},
	{Category: "SANCTIONS", Label: "Sanctions"},
	{Category: "TERRORISM_FINANCING", Label: "Terrorism financing"},
	{Category: "DECENTRALIZED_EXCHANGE", Label: "Decentralized exchange"},
	{Category: "MALWARE", Label: "Malware"},
	{Category: "ILLEGAL_ARMS_TRAFFICKING", Label: "Illegal arms trafficking"},
	{Category: "DRUGS_TRAFFICKING", Label: "Drugs trafficking"},
	{Category: "CHILD_ABUSE_MATERIAL", Label: "Child abuse material"},
	{Category: "PRIVACY_PROTOCOL", Label: "Privacy protocol"},
	{Category: "SANCTIONED_JURISDICTION", Label: "Sanctioned jurisdiction"},
	{Category: "SEIZED_ASSETS", Label: "Seized assets"},
	{Category: "EXTREMISM", Label: "Extremism"},
	{Category: "BRIDGES", Label: "Bridges"},
	{Category: "LENDING_PROTOCOL", Label: "Lending protocol"},
	{Category: "SEIZURE_OF_FUNDS", Label: "Seizure of funds"},
}

var _ aml.SignalCategoryLister = (*Client)(nil)

func (c *Client) SignalCategories() []aml.SignalCategory {
	return signalCategories
}
