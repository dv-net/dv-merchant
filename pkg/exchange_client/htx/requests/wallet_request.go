//nolint:tagliatelle
package requests

type WithdrawVirtualCurrencyRequest struct {
	Address       string `json:"address"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Fee           string `json:"fee,omitempty"`
	Chain         string `json:"chain,omitempty"`
	AddrTag       string `json:"addr-tag,omitempty"`
	ClientOrderID string `json:"client-order-id,omitempty"`
}

type WithdrawalAddressRequest struct {
	Currency string `json:"currency" url:"currency"`
	Chain    string `json:"chain,omitempty" url:"chain,omitempty"`
	Note     string `json:"note,omitempty" url:"note,omitempty"`
	Limit    int    `json:"limit,omitempty" url:"limit,omitempty"`
	FromID   int64  `json:"fromId,omitempty" url:"fromId,omitempty"`
}

type DepositAddressRequest struct {
	Currency string `json:"currency" url:"currency"`
}

type CancelWithdrawalRequest struct {
	WithdrawalTransferID int64 `url:"withdrawal_transfer_id"`
}

type WithdrawalByClientIDRequest struct {
	ClientOrderID string `json:"clientOrderId" url:"clientOrderId"`
}

type TransferType string

func (o TransferType) String() string { return string(o) }

const (
	TransferTypeDeposit  TransferType = "deposit"
	TransferTypeWithdraw TransferType = "withdraw"
)

type WithdrawalDepositHistoryRequest struct {
	Currency string       `url:"currency,omitempty"` // if not specified, all currencies will be returned
	Type     TransferType `url:"type"`               // required field. Either 'deposit' or 'withdraw'
	From     string       `url:"from,omitempty"`     // the transfer id to begin search from
	Size     string       `url:"size,omitempty"`     // number of results to return
	Direct   string       `url:"direct,omitempty"`   // 'prev', 'next' -  ascending, descending
}
