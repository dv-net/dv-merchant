package public_request

type PublicResetPasswordRequest struct {
	Code                 int    `json:"code" param:"code" validate:"required,numeric"`
	Email                string `json:"email" format:"email" validate:"required"`
	Password             string `json:"password" format:"password" validate:"required,eqfield=PasswordConfirmation,min=8,max=32"`
	PasswordConfirmation string `json:"password_confirmation" format:"password" validate:"required,eqfield=Password,min=8,max=32"`
} //	@name	PublicResetPasswordRequest

type PublicAcceptInviteRequest struct {
	Token                string `json:"token" format:"uuid" validate:"required"`
	Email                string `json:"email" format:"email" validate:"required,email"`
	Password             string `json:"password" format:"password" validate:"required,eqfield=PasswordConfirmation,min=8,max=32"`
	PasswordConfirmation string `json:"password_confirmation" format:"password" validate:"required,eqfield=Password,min=8,max=32"`
} //	@name	PublicAcceptInviteRequest

type PublicUserForgotPasswordRequest struct {
	Email string `json:"email" format:"email"`
} //	@name	PublicUserForgotPasswordRequest
