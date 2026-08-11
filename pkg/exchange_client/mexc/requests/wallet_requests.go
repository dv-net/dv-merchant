package requests

type GetDepositAddressRequest struct {
	Coin    string
	Network string
}

type CreateDepositAddressRequest struct {
	Coin    string
	Network string
}

type WithdrawRequest struct {
	Coin    string
	Address string
	Amount  string
	Network string
	Memo    string
}

type GetWithdrawHistoryRequest struct {
	Coin string
}
