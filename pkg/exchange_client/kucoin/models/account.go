//nolint:tagliatelle
package models

import (
	"github.com/shopspring/decimal"
)

type APIKeyInfo struct {
	Remark      string `json:"remark"`
	APIKey      string `json:"apiKey"`
	APIVersion  int    `json:"apiVersion"`
	Permission  string `json:"permission"`
	IPWhitelist string `json:"ipWhitelist"`
	CreatedAt   int64  `json:"createdAt"`
	UID         int64  `json:"uid"`
	IsMaster    bool   `json:"isMaster"`
}

type Account struct {
	ID        string          `json:"id"`
	Currency  string          `json:"currency"`
	Type      AccountType     `json:"type"`
	Balance   decimal.Decimal `json:"balance"`
	Available decimal.Decimal `json:"available"`
	Holds     decimal.Decimal `json:"holds"`
}

type DepositAddress struct {
	Address         string      `json:"address"`
	Memo            string      `json:"memo,omitempty"`
	ChainID         string      `json:"chainId,omitempty"`
	To              AccountType `json:"to"`
	Currency        string      `json:"currency"`
	ContractAddress string      `json:"contractAddress,omitempty"`
	ChainName       string      `json:"chainName"`
}

type WithdrawalHistory struct {
	CurrentPage int64        `json:"currentPage"`
	Items       []Withdrawal `json:"items"`
	PageSize    int64        `json:"pageSize"`
	TotalNum    int64        `json:"totalNum"`
	TotalPage   int64        `json:"totalPage"`
}

type Withdrawal struct {
	ID         string           `json:"id,omitempty"`
	Address    string           `json:"address,omitempty"`
	WalletTxID string           `json:"walletTxId,omitempty"`
	Amount     decimal.Decimal  `json:"amount"`
	Chain      string           `json:"chain,omitempty"`
	Currency   string           `json:"currency,omitempty"`
	Fee        decimal.Decimal  `json:"fee"`
	IsInner    bool             `json:"isInner,omitempty"`
	Memo       string           `json:"memo,omitempty"`
	Remark     string           `json:"remark,omitempty"`
	Status     WithdrawalStatus `json:"status,omitempty"`
	CreatedAt  int64            `json:"createdAt,omitempty"`
	UpdatedAt  int64            `json:"updatedAt,omitempty"`
}
