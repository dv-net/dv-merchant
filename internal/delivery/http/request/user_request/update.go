package user_request

type UpdateRequest struct {
	Location   string `db:"location" json:"location" validate:"required,timezone"`
	Language   string `db:"language" json:"language" validate:"required,min=2,max=2"`
	RateSource string `db:"rate_source" json:"rate_source" validate:"required,oneof=okx htx binance bitget bybit gate dv-min dv-max dv-avg"`
} // @name UpdateUserRequest

type ChangePasswordInternalRequest struct {
	PasswordOld string `json:"password_old" validate:"required,min=8,max=32" format:"password"`
	PasswordNew string `json:"password_new" validate:"required,min=8,max=32" format:"password"`
} // @name ChangePasswordInternalRequest

type ChangePasswordExternalRequestBody struct {
	PasswordNew string `json:"password_new" validate:"required,min=8,max=32" format:"password"`
} // @name ChangePasswordExternalRequest

type ForgotPasswordRequest struct {
	Email string `json:"email" query:"email" validate:"required,email" format:"email"`
} // @name ForgotPasswordRequest
