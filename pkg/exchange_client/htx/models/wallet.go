//nolint:tagliatelle
package models

type WithdrawalAddress struct {
	Currency   string `json:"currency"`
	Chain      string `json:"chain"`
	Note       string `json:"note"`
	AddressTag string `json:"addressTag"`
	Address    string `json:"address"`
}

type DepositAddress struct {
	UserID     int64  `json:"userId"`
	Currency   string `json:"currency"`
	Chain      string `json:"chain"`
	AddressTag string `json:"addressTag,omitempty"`
	Address    string `json:"address"`
}

type CancelWithdrawal struct{}

type WithdrawalByClientID WithdrawalDepositHistory

type WithdrawalDepositHistory struct {
	ID           int           `json:"id"`
	Type         string        `json:"type"`
	Currency     string        `json:"currency"`
	TxHash       string        `json:"tx-hash"`
	Chain        string        `json:"chain"`
	Amount       float64       `json:"amount"`
	Address      string        `json:"address"`
	AddressTag   string        `json:"address-tag"`
	Fee          float64       `json:"fee"`
	State        TransferState `json:"state"`
	ErrorCode    string        `json:"error-code"`
	ErrorMessage string        `json:"error-message"`
	CreatedAt    int64         `json:"created-at"`
	UpdatedAt    int64         `json:"updated-at"`
}

type TransferState string

func (o TransferState) String() string { return string(o) }

const (
	TransferStateVerifying      TransferState = "verifying"
	TransferStateFailed         TransferState = "failed"
	TransferStateSubmitted      TransferState = "submitted"
	TransferStateReexamine      TransferState = "reexamine"
	TransferStateCanceled       TransferState = "canceled"
	TransferStatePass           TransferState = "pass"
	TransferStateReject         TransferState = "reject"
	TransferStatePreTransfer    TransferState = "pre-transfer"
	TransferStateWalletTransfer TransferState = "wallet-transfer"
	TransferStateWalletReject   TransferState = "wallet-reject"
	TransferStateConfirmed      TransferState = "confirmed"
	TransferStateConfirmError   TransferState = "confirm-error"
	TransferStateRepealed       TransferState = "repealed"
)
