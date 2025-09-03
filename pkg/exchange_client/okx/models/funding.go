//nolint:tagliatelle
package models

import "strconv"

type BeneficiaryAccountType int

func (o BeneficiaryAccountType) Int() int { return int(o) }

func (o BeneficiaryAccountType) String() string {
	switch o {
	case BeneficiaryAccountTypeFunding:
		return "funding"
	case BeneficiaryAccountTypeTrading:
		return "trading"
	default:
		return "unknown"
	}
}

const (
	BeneficiaryAccountTypeFunding BeneficiaryAccountType = 6
	BeneficiaryAccountTypeTrading BeneficiaryAccountType = 18
)

type WalletType string

func (o WalletType) String() string { return string(o) }

const (
	WalletTypeExchange WalletType = "exchange"
	WalletTypePrivate  WalletType = "private"
)

type Destination int

func (o Destination) String() string { return strconv.Itoa(int(o)) }
func (o Destination) Int() int       { return int(o) }

const (
	DestinationInternal Destination = 3
	DestinationOnChain  Destination = 4
)

type TransferState string

func (o TransferState) String() string { return string(o) }

const (
	TransferStatePending    TransferState = "Pending withdrawal"
	TransferStateInProgress TransferState = "Withdrawal in progress"
	TransferStateComplete   TransferState = "Withdrawal complete"
	TransferStateCanceled   TransferState = "Cancellation complete"
)

type WithdrawalStatus int

const (
	WithdrawalStatusWaitingWithdrawal WithdrawalStatus = 0
	WithdrawalStatusBroadcasting      WithdrawalStatus = 1
	WithdrawalStatusSuccess           WithdrawalStatus = 2
	WithdrawalStatusWaitingTransfer   WithdrawalStatus = 10
	WithdrawalStatusApproved          WithdrawalStatus = 7
	WithdrawalStatusPendingValidation WithdrawalStatus = 15
	WithdrawalStatusDelayed           WithdrawalStatus = 16
	WithdrawalStatusCanceling         WithdrawalStatus = -3
	WithdrawalStatusCanceled          WithdrawalStatus = -2
	WithdrawalStatusFailed            WithdrawalStatus = -1
)

func (o WithdrawalStatus) String() string { return strconv.Itoa(int(o)) }
func (o WithdrawalStatus) Int() int       { return int(o) }

type (
	Currency struct {
		Ccy                  string `json:"ccy"`
		Name                 string `json:"name"`
		Chain                string `json:"chain"`
		MinDep               string `json:"minDep"`
		MinWd                string `json:"minWd"`
		MaxWd                string `json:"maxWd"`
		Fee                  string `json:"fee"`
		MinFee               string `json:"minFee"`
		MaxFee               string `json:"maxFee"`
		CanDep               bool   `json:"canDep"`
		CanWd                bool   `json:"canWd"`
		CanInternal          bool   `json:"canInternal"`
		MinDepArrivalConfirm string `json:"minDepArrivalConfirm"`
		MinWdUnlockConfirm   string `json:"minWdUnlockConfirm"`
		WdTckSz              string `json:"wdTickSz"`
	}
	FundingBalance struct {
		Ccy       string `json:"ccy"`
		Bal       string `json:"bal"`
		FrozenBal string `json:"frozenBal"`
		AvailBal  string `json:"availBal"`
	}
	FundingTransfer struct {
		TransID string `json:"transId"`
		Ccy     string `json:"ccy"`
		Amt     string `json:"amt"`
		From    int    `json:"from,string"`
		To      int    `json:"to,string"`
	}
	FundingBill struct {
		BillID string `json:"billId"`
		Ccy    string `json:"ccy"`
		Bal    string `json:"bal"`
		BalChg string `json:"balChg"`
		Type   string `json:"type,string"`
		TS     int64  `json:"ts"`
	}
	DepositAddress struct {
		Addr     string `json:"addr"`
		Tag      string `json:"tag,omitempty"`
		Memo     string `json:"memo,omitempty"`
		PmtID    string `json:"pmtId,omitempty"`
		Ccy      string `json:"ccy"`
		Chain    string `json:"chain"`
		CtAddr   string `json:"ctAddr"`
		Selected bool   `json:"selected"`
		To       int    `json:"to,string"`
		TS       int64  `json:"ts"`
	}
	DepositHistory struct {
		Ccy   string `json:"ccy"`
		Chain string `json:"chain"`
		TxID  string `json:"txId"`
		From  string `json:"from"`
		To    string `json:"to"`
		DepID string `json:"depId"`
		Amt   string `json:"amt"`
		State int    `json:"state,string"`
		TS    int64  `json:"ts"`
	}
	Withdrawal struct {
		Ccy   string `json:"ccy"`
		Chain string `json:"chain"`
		WdID  string `json:"wdId"`
		Amt   string `json:"amt"`
	}
	WithdrawalHistory struct {
		Ccy   string           `json:"ccy"`
		Chain string           `json:"chain"`
		TxID  string           `json:"txId"`
		From  string           `json:"from"`
		To    string           `json:"to"`
		Tag   string           `json:"tag,omitempty"`
		PmtID string           `json:"pmtId,omitempty"`
		Memo  string           `json:"memo,omitempty"`
		Amt   string           `json:"amt"`
		Fee   string           `json:"fee"`
		WdID  int64            `json:"wdId,string"`
		State WithdrawalStatus `json:"state,string"`
		TS    int64            `json:"ts,string"`
	}
)
