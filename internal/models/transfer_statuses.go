package models

type TransferStatus string //	@name	TransferStatus

const (
	TransferStatusNew         TransferStatus = "new"
	TransferStatusPending     TransferStatus = "pending"
	TransferStatusProcessing  TransferStatus = "processing"
	TransferStatusInMempool   TransferStatus = "in_mempool"
	TransferStatusUnconfirmed TransferStatus = "unconfirmed"
	TransferStatusCompleted   TransferStatus = "completed"
	TransferStatusFailed      TransferStatus = "failed"
	TransferStatusFrozen      TransferStatus = "frozen"
)

func (t TransferStatus) String() string { return string(t) }
