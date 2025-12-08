package withdrawal_wallets_request

type UpdateAddressesListRequest struct {
	Addresses []WalletAddress `json:"addresses" validate:"omitempty,dive"`
	TOTP      string          `json:"totp" validate:"required,len=6"`
} //	@name	UpdateAddressesListRequest

type WalletAddress struct {
	Address string  `json:"address" validate:"required,max=100,min=1"`
	Name    *string `json:"name" validate:"omitempty,max=100,min=1"`
} //	@name	WalletAddress
