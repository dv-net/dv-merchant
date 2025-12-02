package store_request

type StoreArchiveRequest struct {
	OTP string `json:"otp" validate:"required"`
} //	@name	StoreArchiveRequest
