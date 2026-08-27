package public_request

type RefreshWalletAddressRequest struct {
	Address string `json:"address" validate:"required"`
} //	@name	RefreshWalletAddressRequest

type GetAmlCheckRequest struct {
	Hash string `json:"hash" query:"hash" validate:"required"`
} //	@name	GetAmlCheckRequest
