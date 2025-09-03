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
