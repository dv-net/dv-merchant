package models

type TransferTransactionsStatus string

const (
	TransferTransactionsStatusPending     TransferTransactionsStatus = "pending"
	TransferTransactionsStatusUnconfirmed TransferTransactionsStatus = "unconfirmed"
	TransferTransactionsStatusConfirmed   TransferTransactionsStatus = "confirmed"
	TransferTransactionsStatusFailed      TransferTransactionsStatus = "failed"
)

func (tts TransferTransactionsStatus) String() string {
	return string(tts)
}
