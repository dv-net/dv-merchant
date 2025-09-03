package models

import "strconv"

type Network struct {
	AddressRegex            string `json:"addressRegex"`
	Coin                    string `json:"coin"`
	DepositDesc             string `json:"depositDesc,omitempty"`
	DepositEnable           bool   `json:"depositEnable"`
	IsDefault               bool   `json:"isDefault"`
	MemoRegex               string `json:"memoRegex"`
	MinConfirm              int    `json:"minConfirm"`
	DepositDust             string `json:"depositDust"`
	Name                    string `json:"name"`
	Network                 string `json:"network"`
	SpecialTips             string `json:"specialTips"`
	UnLockConfirm           int    `json:"unLockConfirm"`
	WithdrawDesc            string `json:"withdrawDesc,omitempty"`
	WithdrawEnable          bool   `json:"withdrawEnable"`
	WithdrawFee             string `json:"withdrawFee"`
	WithdrawIntegerMultiple string `json:"withdrawIntegerMultiple"`
	WithdrawMax             string `json:"withdrawMax"`
	WithdrawMin             string `json:"withdrawMin"`
	SameAddress             bool   `json:"sameAddress"`
	EstimatedArrivalTime    int    `json:"estimatedArrivalTime"`
	Busy                    bool   `json:"busy"`
	ContractAddressUrl      string `json:"contractAddressUrl"`
	ContractAddress         string `json:"contractAddress"`
}

type CoinInfo struct {
	Coin              string                   `json:"coin"`
	DepositAllEnable  bool                     `json:"depositAllEnable"`
	Free              string                   `json:"free"`
	Freeze            string                   `json:"freeze"`
	Ipoable           string                   `json:"ipoable"`
	Ipoing            string                   `json:"ipoing"`
	IsLegalMoney      bool                     `json:"isLegalMoney"`
	Locked            string                   `json:"locked"`
	Name              string                   `json:"name"`
	NetworkList       []Network                `json:"networkList"`
	Storage           string                   `json:"storage"`
	Trading           bool                     `json:"trading"`
	WithdrawAllEnable bool                     `json:"withdrawAllEnable"`
	Withdrawing       string                   `json:"withdrawing"`
	Filters           []map[string]interface{} `json:"filters"`
}

type WithdrawalStatus int

func (o WithdrawalStatus) String() string { return strconv.Itoa(int(o)) }
func (o WithdrawalStatus) Int() int       { return int(o) }

const (
	WithdrawalStatusEmailSent WithdrawalStatus = iota
	WithdrawalStatusAwaitingApproval
	WithdrawalStatusRejected
	WithdrawalStatusProcessing
	WithdrawalStatusCompleted
)

type WithdrawalInfo struct {
	Id              string           `json:"id"`
	Amount          string           `json:"amount"`
	TransactionFee  string           `json:"transactionFee"`
	Coin            string           `json:"coin"`
	Status          WithdrawalStatus `json:"status"`
	Address         string           `json:"address"`
	TxId            string           `json:"txId"`
	ApplyTime       string           `json:"applyTime"`
	Network         string           `json:"network"`
	TransferType    int              `json:"transferType"`
	WithdrawOrderId string           `json:"withdrawOrderId"`
	Info            string           `json:"info,omitempty"`
	ConfirmNo       int              `json:"confirmNo"`
	WalletType      int              `json:"walletType"`
	TxKey           string           `json:"txKey,omitempty"`
	CompleteTime    string           `json:"completeTime"`
}

type WithdrawalWallet struct {
	Address     string `json:"address"`
	AddressTag  string `json:"addressTag"`
	Coin        string `json:"coin"`
	Name        string `json:"name"`
	Network     string `json:"network"`
	Origin      string `json:"origin"`
	OriginType  string `json:"originType"`
	WhiteStatus bool   `json:"whiteStatus"`
}

type TransferType string

func (o TransferType) String() string {
	return string(o)
}

const (
	TransferTypeSpotToFunding TransferType = "MAIN_FUNDING"
	TransferTypeFundingToSpot TransferType = "FUNDING_MAIN"
)

type AssetBalance struct {
	Asset        string `json:"asset"`
	Free         string `json:"free"`
	Locked       string `json:"locked"`
	Freeze       string `json:"freeze"`
	Withdrawing  string `json:"withdrawing"`
	BtcValuation string `json:"btcValuation"`
}
