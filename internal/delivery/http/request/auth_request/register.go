package auth_request

type RegisterRequest struct {
	Email       string `db:"email" json:"email" validate:"required,email"`
	Password    string `db:"password" json:"password" validate:"required,min=8,max=32"`
	Location    string `db:"location" json:"location" validate:"required,timezone"`
	Language    string `db:"language" json:"language" validate:"required,min=2,max=2"`
	CfTurnstile string `json:"cf-turnstile-response"` //nolint:tagliatelle
	Mnemonic    string `json:"mnemonic"`
} //	@name	RegisterRequest
