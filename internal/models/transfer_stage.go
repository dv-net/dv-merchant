package models

type TransferStage string //	@name	TransferStage

func (t TransferStage) String() string {
	return string(t)
}

const (
	TransferStageInProgress TransferStage = "in_progress"
	TransferStageFailed     TransferStage = "failed"
	TransferStageCompleted  TransferStage = "completed"
)

func ResolveTransferStageByStatus(status TransferStatus) TransferStage {
	switch status {
	case TransferStatusFailed:
		return TransferStageFailed
	case TransferStatusCompleted:
		return TransferStageCompleted
	default:
		return TransferStageInProgress
	}
}
