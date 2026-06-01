package public_request

type RefreshWalletAddressRequest struct {
	Address string `json:"address" validate:"required"`
} //	@name	RefreshWalletAddressRequest
