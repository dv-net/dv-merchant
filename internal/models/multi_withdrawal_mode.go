package models

type MultiWithdrawalMode string

const (
	MultiWithdrawalModeRandom     MultiWithdrawalMode = "random"
	MultiWithdrawalModeDisabled   MultiWithdrawalMode = "disabled"
	MultiWithdrawalModeProcessing MultiWithdrawalMode = "processing"
	MultiWithdrawalModeManual     MultiWithdrawalMode = "manual"
)

func (w MultiWithdrawalMode) String() string {
	return string(w)
}

func (w MultiWithdrawalMode) IsValidByCurrency(curr Currency) bool {
	switch w {
	case MultiWithdrawalModeDisabled:
		return true
	default:
		return curr.Blockchain.IsBitcoinLike()
	}
}
