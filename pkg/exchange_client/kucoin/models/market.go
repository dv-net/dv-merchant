//nolint:tagliatelle
package models

import "github.com/shopspring/decimal"

type Currency struct {
	Currency        string   `json:"currency"`
	Name            string   `json:"name"`
	FullName        string   `json:"fullName"`
	Precision       int      `json:"precision"`
	Confirms        int      `json:"confirms,omitempty"`
	ContractAddress string   `json:"contractAddress,omitempty"`
	Chains          []*Chain `json:"chains,omitempty"`
}

type Chain struct {
	ChainName           string          `json:"chainName"`
	WithdrawalMinFee    decimal.Decimal `json:"withdrawalMinFee"`
	WithdrawalMinSize   decimal.Decimal `json:"withdrawalMinSize"`
	WithdrawFeeRate     decimal.Decimal `json:"withdrawFeeRate"`
	WithdrawPrecision   int             `json:"withdrawPrecision"`
	DepositMinSize      decimal.Decimal `json:"depositMinSize,omitempty"`
	IsWithdrawalEnabled bool            `json:"isWithdrawalEnabled"`
	IsDepositEnabled    bool            `json:"isDepositEnabled"`
	PreConfrims         int             `json:"preConfirms"`
	ContractAddress     string          `json:"contractAddress,omitempty"`
	ChainID             string          `json:"chainId"`
	Confirms            int             `json:"confirms"`
}
