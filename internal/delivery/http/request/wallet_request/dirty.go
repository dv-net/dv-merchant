package wallet_request

type MarkIsDirtyRequest struct {
	Address string `json:"address" validate:"required"`
} //	@name	MarkIsDirtyRequest
