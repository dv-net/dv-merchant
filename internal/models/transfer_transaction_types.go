package models

type TransferTransactionType string

const (
	TransferTransactionTypeDelegateResources TransferTransactionType = "resource_delegation"
	TransferTransactionTypeReclaimResources  TransferTransactionType = "resource_reclaim"
	TransferTransactionTypeSendBurnBaseAsset TransferTransactionType = "send_burn_base_asset"
	TransferTransactionTypeAccountActivation TransferTransactionType = "account_activation"
	TransferTransactionTypeTransfer          TransferTransactionType = "transfer"
)

func (t TransferTransactionType) String() string {
	return string(t)
}
