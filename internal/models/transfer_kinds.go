package models

type TransferKind string //	@name	TransferKind

const (
	TransferKindFromAddress    TransferKind = "from_address"
	TransferKindFromProcessing TransferKind = "from_processing"
)

func (o TransferKind) String() string {
	return string(o)
}
