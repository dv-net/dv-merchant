package responses

import "github.com/shopspring/decimal"

const (
	WithdrawalStatusApply      int64 = 1
	WithdrawalStatusAuditing   int64 = 2
	WithdrawalStatusProcessing int64 = 4
	WithdrawalStatusRejected   int64 = 5
	WithdrawalStatusCompleted  int64 = 6
	WithdrawalStatusFailed     int64 = 7
)

type CoinNetwork struct {
	Name              string          `json:"name"`
	Network           string          `json:"network"`
	IsDefault         bool            `json:"isDefault"`
	MinConfirm        int64           `json:"minConfirm"`
	WithdrawEnable    bool            `json:"withdrawEnable"`
	DepositEnable     bool            `json:"depositEnable"`
	WithdrawFee       decimal.Decimal `json:"withdrawFee"`
	WithdrawMax       decimal.Decimal `json:"withdrawMax"`
	WithdrawMin       decimal.Decimal `json:"withdrawMin"`
	DepositMin        decimal.Decimal `json:"depositMin"`
	WithdrawPrecision int32           `json:"withdrawPrecision"`
	DepositPrecision  int32           `json:"depositPrecision"`
	ContractAddress   string          `json:"contractAddress"`
	NeedTagOrMemo     string          `json:"needTagOrMemo"`
	DisplayName       string          `json:"displayName"`
}

func (o CoinNetwork) RequiresTag() bool {
	return o.NeedTagOrMemo == "true"
}

type CoinConfig struct {
	Coin        string        `json:"coin"`
	Name        string        `json:"name"`
	NetworkList []CoinNetwork `json:"networkList"`
}

type GetCoinsConfigResponse []CoinConfig

type DepositAddress struct {
	CoinID            int64  `json:"coinId"`
	Coin              string `json:"coin"`
	Network           string `json:"network"`
	Address           string `json:"address"`
	AddressWithPrefix string `json:"addressWithPrefix"`
	Tag               string `json:"tag"`
	WalletType        int    `json:"walletType"`
}

type GetDepositAddressResponse struct {
	Data  []DepositAddress `json:"data"`
	Total int              `json:"total"`
}

type WithdrawResponse struct {
	ID string `json:"id"`
}

type WithdrawHistoryItem struct {
	ID              string          `json:"id"`
	Coin            string          `json:"coin"`
	Network         string          `json:"network"`
	Address         string          `json:"address"`
	SourceAddress   string          `json:"sourceAddress"`
	Amount          decimal.Decimal `json:"amount"`
	TransactionFee  decimal.Decimal `json:"transactionFee"`
	Status          int64           `json:"status"`
	TxID            string          `json:"txId"`
	ApplyTime       string          `json:"applyTime"`
	ConfirmNo       int64           `json:"confirmNo"`
	TransferType    int64           `json:"transferType"`
	WithdrawOrderID string          `json:"withdrawOrderId"`
	Info            string          `json:"info"`
}

type GetWithdrawHistoryResponse []WithdrawHistoryItem
