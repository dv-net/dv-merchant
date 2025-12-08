package user_request

type ConfirmEmailRequest struct {
	Code int `json:"code" query:"code" validate:"required"`
} //	@name	ConfirmEmailRequest

type ChangeEmailRequest struct {
	NewEmail             string `json:"new_email" query:"new_email" validate:"required,email"`
	NewEmailConfirmation string `json:"new_email_confirmation" query:"new_email_confirmation"`
	Code                 string `json:"code" query:"code" validate:"omitempty,numeric"`
} //	@name	ChangeEmailRequest
