package two_factor_auth

type ChangeStatus struct {
	OTP string `json:"otp" validate:"required"`
} // @name ChangeTwoFactorAuthStatusRequest
