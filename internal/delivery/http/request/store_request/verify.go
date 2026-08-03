package store_request

type ResendStoreVerificationRequest struct {
	Comment string `json:"comment" validate:"omitempty,min=1,max=255"`
} //	@name	ResendStoreVerificationRequest
