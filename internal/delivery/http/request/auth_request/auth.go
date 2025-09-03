package auth_request

type AuthRequest struct {
	Email      string `db:"email" json:"email" validate:"required,email"`
	Password   string `db:"password" json:"password" validate:"required,min=8,max=32"`
	RememberMe bool   `db:"remember" json:"remember_me"`
} // @name AuthRequest
