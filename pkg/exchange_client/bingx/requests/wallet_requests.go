package requests

type GetDepositAddressRequest struct {
	Coin   string
	Offset int
	Limit  int
}

type WithdrawRequest struct {
	Coin       string
	Network    string
	Address    string
	AddressTag string
	Amount     string
	WalletType int
}

type GetWithdrawHistoryRequest struct {
	Coin string
}

type TransferRequest struct {
	Type   string
	Asset  string
	Amount string
}
