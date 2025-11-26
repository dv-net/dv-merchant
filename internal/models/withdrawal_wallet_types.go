package models

type WithdrawalWalletType string //	@name	WithdrawalWalletType

const (
	WithdrawalTypeBalanceLimit     WithdrawalWalletType = "balance"
	WithdrawalTypeInterval         WithdrawalWalletType = "interval"
	WithdrawalTypeManual           WithdrawalWalletType = "manual"
	WithdrawalTypeDeposit          WithdrawalWalletType = "deposit"
	WithdrawalTypeLimitAndInterval WithdrawalWalletType = "limit_and_balance"
)
