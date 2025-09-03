//nolint:tagliatelle
package models

import "github.com/shopspring/decimal"

type AccountAsset struct {
	Coin           string          `json:"coin"`
	Available      decimal.Decimal `json:"available"`
	Frozen         decimal.Decimal `json:"frozen"`
	Locked         decimal.Decimal `json:"locked"`
	LimitAvailable decimal.Decimal `json:"limitAvailable"`
	UTime          string          `json:"uTime"`
}

type DepositRecord struct {
	OrderID  string  `json:"orderId"`
	TradeID  string  `json:"tradeId"`
	Coin     string  `json:"coin"`
	Type     string  `json:"type"`
	Size     float64 `json:"size,string"`
	Status   string  `json:"status"`
	ToAddr   string  `json:"toAddress"`
	Dest     string  `json:"dest"`
	Chain    string  `json:"chain"`
	FromAddr string  `json:"fromAddress"`
	CTime    int64   `json:"cTime,string"`
	UTime    int64   `json:"uTime,string"`
}

type WithdrawalRecord struct {
	OrderID   string           `json:"orderId"`
	TradeID   string           `json:"tradeId"`
	Coin      string           `json:"coin"`
	Dest      string           `json:"dest"`
	ClientOid string           `json:"clientOid"`
	Type      string           `json:"type"`
	Tag       string           `json:"tag,omitempty"`
	Size      decimal.Decimal  `json:"size"`
	Fee       decimal.Decimal  `json:"fee"`
	Status    WithdrawalStatus `json:"status"`
	ToAddress string           `json:"toAddress"`
	FromAddr  string           `json:"fromAddress"`
	Confirm   int64            `json:"confirm,string"`
	Chain     string           `json:"chain"`
	CTime     int64            `json:"cTime,string"`
	UTime     int64            `json:"uTime,string"`
}

type WithdrawalStatus string

func (o WithdrawalStatus) String() string { return string(o) }

const (
	WithdrawalStatusFailed  WithdrawalStatus = "failed"
	WithdrawalStatusSuccess WithdrawalStatus = "success"
	WithdrawalStatusPending WithdrawalStatus = "pending"
)

type Withdrawal struct {
	OrderID   string `json:"orderId"`
	ClientOid string `json:"clientOid"`
}
