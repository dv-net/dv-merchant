package models

type WithdrawalStatus string //	@name	WithdrawalStatus

const (
	WithdrawalStatusEnabled  WithdrawalStatus = "enabled"
	WithdrawalStatusDisabled WithdrawalStatus = "disabled"
	// WithdrawalStatusSuspended WithdrawalStatus = "suspended"
)

var validStatuses = map[WithdrawalStatus]struct{}{
	WithdrawalStatusEnabled:  {},
	WithdrawalStatusDisabled: {},
	// WithdrawalStatusSuspended: {},
}

func (o WithdrawalStatus) IsValid() bool {
	_, ok := validStatuses[o]
	return ok
}

func (o WithdrawalStatus) String() string {
	return string(o)
}
