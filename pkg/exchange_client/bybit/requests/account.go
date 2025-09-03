//nolint:tagliatelle
package requests

import "github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/models"

type GetTranscationLogRequest struct {
	Category string `json:"category,omitempty" url:"category,omitempty"`
	Currency string `json:"currency,omitempty" url:"currency,omitempty"`
	BaseCoin string `json:"baseCoin,omitempty" url:"baseCoin,omitempty"`
}

type GetTradingBalanceRequest struct {
	AccountType string `json:"accountType,omitempty" url:"accountType,omitempty"`
	Coin        string `json:"coin,omitempty" url:"coin,omitempty"`
}

type GetFundingBalanceRequest struct {
	AccountType string `json:"accountType" url:"accountType"`
	Coin        string `json:"coin,omitempty" url:"coin,omitempty"`
}

type GetDepositAddressRequest struct {
	Coin      string `json:"coin" url:"coin"`
	ChainType string `json:"chainType,omitempty" url:"chainType,omitempty"`
}

type CreateInternalTransferRequest struct {
	TransferID      string             `json:"transferId" url:"transferId"`
	Coin            string             `json:"coin" url:"coin"`
	Amount          string             `json:"amount" url:"amount"`
	FromAccountType models.AccountType `json:"fromAccountType" url:"fromAccountType"`
	ToAccountType   models.AccountType `json:"toAccountType" url:"toAccountType"`
}

type CreateWithdrawRequest struct {
	Coin       string `json:"coin"`
	Chain      string `json:"chain,omitempty"`
	Address    string `json:"address"`
	Tag        string `json:"tag,omitempty"`
	Amount     string `json:"amount"`
	Timestamp  string `json:"timestamp"`
	ForceChain string `json:"forceChain,omitempty"`
	FeeType    string `json:"feeType,omitempty"`
}

type GetWithdrawRequest struct {
	WithdrawID   string `json:"withdrawId,omitempty" url:"withdrawId,omitempty"`
	TxID         string `json:"txId,omitempty" url:"txId,omitempty"`
	Coin         string `json:"coin,omitempty" url:"coin,omitempty"`
	WithdrawType string `json:"withdrawType,omitempty" url:"withdrawType,omitempty"`
	StartTime    int64  `json:"startTime,omitempty" url:"startTime,omitempty"`
	EndTime      int64  `json:"endTime,omitempty" url:"endTime,omitempty"`
	Limit        int    `json:"limit,omitempty" url:"limit,omitempty"`
}
