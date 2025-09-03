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
