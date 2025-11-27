package setting_request

type CreateRequest struct {
	Name  string  `json:"name" validate:"required,min=2,max=255"`
	Value *string `json:"value" validate:"omitempty,min=0,max=255"`
	OTP   *string `json:"otp" validate:"omitempty"`
} //	@name	CreateSettingRequest
