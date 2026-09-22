package models

type AmlRiskFlag string

const (
	AmlRiskFlagSanctions        AmlRiskFlag = "sanctions"
	AmlRiskFlagDarknetIllicit   AmlRiskFlag = "darknet_illicit"
	AmlRiskFlagMixerPrivacy     AmlRiskFlag = "mixer_privacy"
	AmlRiskFlagHighRiskExchange AmlRiskFlag = "high_risk_exchange"
)

func (f AmlRiskFlag) IsValid() bool {
	switch f {
	case AmlRiskFlagSanctions, AmlRiskFlagDarknetIllicit, AmlRiskFlagMixerPrivacy, AmlRiskFlagHighRiskExchange:
		return true
	default:
		return false
	}
}
