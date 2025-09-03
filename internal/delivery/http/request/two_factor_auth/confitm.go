package two_factor_auth

type Confirm struct {
	OTP string `json:"otp" validate:"required"`
} // @name ConfirmTwoFactorAuthRequest
