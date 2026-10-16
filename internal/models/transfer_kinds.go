package models

type TransferKind string //	@name	TransferKind

const (
	TransferKindFromAddress    TransferKind = "from_address"
	TransferKindFromProcessing TransferKind = "from_processing"
	// TransferKindRefund is a refund of a blocked deposit from the hot wallet to an external address
	TransferKindRefund TransferKind = "refund"
)

func (o TransferKind) String() string {
	return string(o)
}
