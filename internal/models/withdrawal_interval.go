package models

type WithdrawalInterval string //	@name	WithdrawalInterval

const (
	WithdrawalIntervalNever        WithdrawalInterval = "never"
	WithdrawalIntervalEveryOneMin  WithdrawalInterval = "every-one-min"
	WithdrawalIntervalEvery12hours WithdrawalInterval = "every-12hours"
	WithdrawalIntervalEveryDay     WithdrawalInterval = "every-day"
	WithdrawalIntervalEvery3Days   WithdrawalInterval = "every-3days"
	WithdrawalIntervalEveryWeek    WithdrawalInterval = "every-week"
)

var validIntervals = map[WithdrawalInterval]struct{}{
	WithdrawalIntervalNever:        {},
	WithdrawalIntervalEveryOneMin:  {},
	WithdrawalIntervalEvery12hours: {},
	WithdrawalIntervalEveryDay:     {},
	WithdrawalIntervalEvery3Days:   {},
	WithdrawalIntervalEveryWeek:    {},
}

// IsValid checks if the WithdrawalInterval is valid.
func (wi WithdrawalInterval) IsValid() bool {
	_, ok := validIntervals[wi]
	return ok
}

func (wi WithdrawalInterval) String() string {
	return string(wi)
}
