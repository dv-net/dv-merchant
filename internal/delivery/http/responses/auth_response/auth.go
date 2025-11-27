package auth_response

import "github.com/dv-net/dv-merchant/internal/service/user"

type RegisterRootResponse struct {
	Token string `json:"token"`
} //	@name	RegisterRootResponse

type RegisterUserResponse struct {
	UserInfo *user.RegisterUserDTO `json:"user_info"`
	Token    string                `json:"token"`
} //	@name	RegisterUserResponse

type AuthResponse struct {
	Token string `json:"token"`
} //	@name	AuthResponse
