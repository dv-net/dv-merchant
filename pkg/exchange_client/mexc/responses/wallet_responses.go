//nolint:tagliatelle
package responses

import "github.com/shopspring/decimal"

type CoinNetwork struct {
	Coin                    string          `json:"coin"`
	DepositDesc             string          `json:"depositDesc"`
	DepositEnable           bool            `json:"depositEnable"`
	MinConfirm              int64           `json:"minConfirm"`
	DepositPreConfirms      int64           `json:"depositPreConfirms"`
	Name                    string          `json:"name"`
	Network                 string          `json:"network"`
	WithdrawEnable          bool            `json:"withdrawEnable"`
	WithdrawFee             decimal.Decimal `json:"withdrawFee"`
	WithdrawIntegerMultiple decimal.Decimal `json:"withdrawIntegerMultiple"`
	WithdrawMax             decimal.Decimal `json:"withdrawMax"`
	WithdrawMin             decimal.Decimal `json:"withdrawMin"`
	SameAddress             bool            `json:"sameAddress"`
	Contract                string          `json:"contract"`
	WithdrawTips            *string         `json:"withdrawTips"`
	DepositTips             *string         `json:"depositTips"`
	NetWork                 string          `json:"netWork"`
}

type CoinConfig struct {
	Coin        string        `json:"coin"`
	Name        string        `json:"name"`
	NetworkList []CoinNetwork `json:"networkList"`
}

type GetCoinsConfigResponse []CoinConfig

type DepositAddress struct {
	Coin    string `json:"coin"`
	NetWork string `json:"netWork"`
	Address string `json:"address"`
	Memo    string `json:"memo"`
}

type GetDepositAddressResponse []DepositAddress

type CreateDepositAddressResponse DepositAddress

const (
	WithdrawalStatusApply         int64 = 1
	WithdrawalStatusAuditing      int64 = 2
	WithdrawalStatusWait          int64 = 3
	WithdrawalStatusProcessing    int64 = 4
	WithdrawalStatusWaitPackaging int64 = 5
	WithdrawalStatusWaitConfirm   int64 = 6
	WithdrawalStatusSuccess       int64 = 7
	WithdrawalStatusFailed        int64 = 8
	WithdrawalStatusCancel        int64 = 9
	WithdrawalStatusManual        int64 = 10
)

type WithdrawResponse struct {
	ID string `json:"id"`
}

type WithdrawHistoryItem struct {
	ID             string          `json:"id"`
	TxID           string          `json:"txId"`
	Coin           string          `json:"coin"`
	Network        string          `json:"network"`
	Address        string          `json:"address"`
	Amount         decimal.Decimal `json:"amount"`
	Status         int64           `json:"status"`
	ApplyTime      int64           `json:"applyTime"`
	TransactionFee decimal.Decimal `json:"transactionFee"`
	ConfirmNo      int64           `json:"confirmNo"`
	Memo           string          `json:"memo"`
}

type GetWithdrawHistoryResponse []WithdrawHistoryItem
