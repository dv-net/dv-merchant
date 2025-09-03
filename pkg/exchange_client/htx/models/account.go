//nolint:tagliatelle
package models

type MarketStatus int

const (
	MarketStatusNormal    MarketStatus = 1
	MarketStatusSuspended MarketStatus = 2
	MarketStatusCancelled MarketStatus = 3
)

type MarketStatusResponse struct {
	MarketStatus    MarketStatus `json:"marketStatus,omitempty"`
	HaltStartTime   int64        `json:"haltStartTime,omitempty"`
	HaltEndTime     int64        `json:"haltEndTime,omitempty"`
	HaltReason      int          `json:"haltReason,omitempty"`
	AffectedSymbols string       `json:"affectedSymbols,omitempty"`
}

type AccountState string

func (o AccountState) String() string { return string(o) }

const (
	AccountStateLocked  AccountState = "lock"
	AccountStateWorking AccountState = "working"
)

type AccountType string

func (o AccountType) String() string { return string(o) }

const (
	AccountTypeSpot   AccountType = "spot"
	AccountTypeMargin AccountType = "margin"
	AccountTypeOTC    AccountType = "otc"
	AccountTypePoint  AccountType = "point"
	AccountTypeSuper  AccountType = "super-margin"
	AccountTypeInvest AccountType = "investment"
)

type Account struct {
	ID      int64        `json:"id"`
	State   AccountState `json:"state"`
	Type    AccountType  `json:"type"`
	SubType string       `json:"subtype,omitempty"`
}

type BalanceType string

func (o BalanceType) String() string { return string(o) }

const (
	BalanceTypeTrade    BalanceType = "trade"
	BalanceTypeFrozen   BalanceType = "frozen"
	BalanceTypeLoan     BalanceType = "loan"
	BalanceTypeInterest BalanceType = "interest"
	BalanceTypeLock     BalanceType = "lock"
)

type ListItem struct {
	Balance   string      `json:"balance,omitempty"`
	Currency  string      `json:"currency,omitempty"`
	Available string      `json:"available,omitempty"`
	Type      BalanceType `json:"type,omitempty"`
	SeqNum    string      `json:"seq-num,omitempty"`
	Debt      string      `json:"debt,omitempty"`
}

type AccountBalance struct {
	ID    int64        `json:"id,omitempty"`
	Type  AccountType  `json:"type,omitempty"`
	State AccountState `json:"state,omitempty"`
	List  []ListItem   `json:"list,omitempty"`
}

type AssetValuation struct {
	Balance   string `json:"balance,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}
