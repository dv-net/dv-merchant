//nolint:tagliatelle
package okx

type Response[T any] struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type MaintenanceState string

const (
	MaintenanceStateScheduled MaintenanceState = "scheduled"
	MaintenanceStateOngoing   MaintenanceState = "ongoing"
	MaintenanceStatePreOpen   MaintenanceState = "pre_open"
	MaintenanceStateCompleted MaintenanceState = "completed"
	MaintenanceStateCanceled  MaintenanceState = "canceled"
)

func (o MaintenanceState) String() string { return string(o) }

type StatusResponse struct {
	State MaintenanceState `json:"state,omitempty"`
}

type TradingAccountBalanceData struct {
	Details []TradingAccountBalanceDetails `json:"details,omitempty"`
	TotalEq string                         `json:"totalEq,omitempty"`
}

type TradingAccountBalanceDetails struct {
	AvailBal string `json:"availBal,omitempty"`
	Ccy      string `json:"ccy,omitempty"`
	Eq       string `json:"eq,omitempty"`
	EqUsd    string `json:"eqUsd,omitempty"`
}

type FundingAccountBalance struct {
	Ccy       string `json:"ccy"`
	Bal       string `json:"bal"`
	FrozenBal string `json:"frozenBal"`
	AvailBal  string `json:"availBal"`
}
