package two_factor_response

type GetTwoFactorAuthSecretResponse struct {
	Secret      string `json:"secret,omitempty"`
	IsConfirmed bool   `json:"is_confirmed"`
} // @name GetTwoFactorAuthDataResponse
