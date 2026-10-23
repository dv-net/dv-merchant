//nolint:tagliatelle
package responses

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

const (
	WithdrawStatusInitial = "Initial"
	WithdrawStatusPending = "Pending"
	WithdrawStatusSettled = "Settled"
	WithdrawStatusSuccess = "Success"
	WithdrawStatusFailure = "Failure"
)

const (
	WithdrawStatusPropCancelPending = "cancel-pending"
	WithdrawStatusPropCanceled      = "canceled"
	WithdrawStatusPropCancelDenied  = "cancel-denied"
	WithdrawStatusPropReturn        = "return"
	WithdrawStatusPropOnHold        = "onhold"
)

// Limit is a Kraken amount that is `false` when there is no limit.
type Limit struct {
	Value   decimal.Decimal
	Limited bool
}

func (l *Limit) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if bytes.Equal(data, []byte("false")) || bytes.Equal(data, []byte("null")) {
		*l = Limit{}
		return nil
	}
	if err := l.Value.UnmarshalJSON(data); err != nil {
		return fmt.Errorf("kraken limit: %w", err)
	}
	l.Limited = true
	return nil
}

// UnixTime accepts both 1688665496 and "1688665496".
type UnixTime int64

func (t *UnixTime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(bytes.TrimSpace(data)), `"`)
	if s == "" || s == "null" {
		*t = 0
		return nil
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("kraken time: %w", err)
	}
	*t = UnixTime(n)
	return nil
}

type DepositMethod struct {
	Method          string          `json:"method"`
	Limit           Limit           `json:"limit"`
	Fee             decimal.Decimal `json:"fee"`
	AddressSetupFee decimal.Decimal `json:"address-setup-fee"`
	GenAddress      bool            `json:"gen-address"`
	Minimum         decimal.Decimal `json:"minimum"`
}

type DepositMethodsResponse []DepositMethod

func (r DepositMethodsResponse) Find(method string) (DepositMethod, bool) {
	for _, m := range r {
		if m.Method == method {
			return m, true
		}
	}
	return DepositMethod{}, false
}

type DepositAddress struct {
	Address  string   `json:"address"`
	ExpireTm UnixTime `json:"expiretm"` // 0 means the address does not expire
	New      bool     `json:"new"`
	Memo     string   `json:"memo,omitempty"`
	Tag      string   `json:"tag,omitempty"`
}

type DepositAddressesResponse []DepositAddress

type WithdrawMethod struct {
	Asset   string          `json:"asset"`
	Method  string          `json:"method"`
	Network string          `json:"network"`
	Minimum decimal.Decimal `json:"minimum"`
}

type WithdrawMethodsResponse []WithdrawMethod

type WithdrawAddress struct {
	Address  string `json:"address"`
	Asset    string `json:"asset"`
	Method   string `json:"method"`
	Key      string `json:"key"`
	Memo     string `json:"memo,omitempty"`
	Tag      string `json:"tag,omitempty"`
	Verified bool   `json:"verified"`
}

type WithdrawAddressesResponse []WithdrawAddress

// FindByAddress returns the verified saved address matching address and method.
// EVM addresses are compared case-insensitively, others exactly.
// Empty method matches any method.
func (r WithdrawAddressesResponse) FindByAddress(address, method string) (WithdrawAddress, bool) {
	for _, a := range r {
		if !a.Verified || !sameAddress(a.Address, address) {
			continue
		}
		if method != "" && a.Method != method {
			continue
		}
		return a, true
	}
	return WithdrawAddress{}, false
}

func sameAddress(a, b string) bool {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if strings.HasPrefix(a, "0x") && strings.HasPrefix(b, "0x") {
		return strings.EqualFold(a, b)
	}
	return a == b
}

type WithdrawInfoResponse struct {
	Method string          `json:"method"`
	Limit  decimal.Decimal `json:"limit"`  // max amount that can be withdrawn now
	Amount decimal.Decimal `json:"amount"` // amount that will be received after fee
	Fee    decimal.Decimal `json:"fee"`
}

type WithdrawResponse struct {
	RefID string `json:"refid"`
}

type Withdrawal struct {
	Method     string          `json:"method"`
	Network    string          `json:"network"`
	AClass     string          `json:"aclass"`
	Asset      string          `json:"asset"`
	RefID      string          `json:"refid"`
	TxID       string          `json:"txid"`
	Info       string          `json:"info"` // destination address
	Amount     decimal.Decimal `json:"amount"`
	Fee        decimal.Decimal `json:"fee"`
	Time       UnixTime        `json:"time"`
	Status     string          `json:"status"`
	StatusProp string          `json:"status-prop"`
	Key        string          `json:"key"`
}

func (w Withdrawal) IsSuccess() bool {
	return w.Status == WithdrawStatusSuccess && !w.isCanceledOrReturned()
}

func (w Withdrawal) IsFailed() bool {
	return w.Status == WithdrawStatusFailure || w.isCanceledOrReturned()
}

func (w Withdrawal) IsInProgress() bool {
	return !w.IsSuccess() && !w.IsFailed()
}

func (w Withdrawal) isCanceledOrReturned() bool {
	return w.StatusProp == WithdrawStatusPropCanceled || w.StatusProp == WithdrawStatusPropReturn
}

type WithdrawStatusResponse []Withdrawal

func (r WithdrawStatusResponse) FindByRefID(refID string) (Withdrawal, bool) {
	for _, w := range r {
		if w.RefID == refID {
			return w, true
		}
	}
	return Withdrawal{}, false
}
