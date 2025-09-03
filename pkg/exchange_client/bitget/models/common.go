//nolint:tagliatelle
package models

type (
	AccountBalance struct {
		AccountType string `json:"accountType"`
		BalanceUSDT string `json:"usdtBalance"`
	}
	DepositAddress struct {
		Address string `json:"address"`
		Chain   string `json:"chain"`
		Coin    string `json:"coin"`
		Tag     string `json:"tag,omitempty"`
		URL     string `json:"url,omitempty"`
	}
	ServerTime struct {
		ServerTime int64 `json:"server_time,string"`
	}
)

type (
	TransferType string
)

const (
	TransferTypeOnChain  TransferType = "on_chain"
	TransferTypeInternal TransferType = "internal_transfer"
)
