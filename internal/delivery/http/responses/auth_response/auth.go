package auth_response

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/user_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/user"
)

type RegisterRootResponse struct {
	Token string `json:"token"`
} //	@name	RegisterRootResponse

// RegisterUserInfo is an explicit projection of the registered user.
// Never embed models.User directly: it carries the password hash, remember_token and dvnet_token.
type RegisterUserInfo struct {
	User      *user_response.AuthUserResponse `json:"user"`
	OwnerInfo *processing.RegisterOwnerInfo   `json:"owner_info"`
} //	@name	RegisterUserInfo

type RegisterUserResponse struct {
	UserInfo *RegisterUserInfo `json:"user_info"`
	Token    string            `json:"token"`
} //	@name	RegisterUserResponse

func NewRegisterUserInfo(info *user.RegisterUserDTO, roles []models.UserRole) *RegisterUserInfo {
	if info == nil {
		return nil
	}

	res := &RegisterUserInfo{OwnerInfo: info.OwnerInfo}
	if info.User != nil {
		res.User = &user_response.AuthUserResponse{}
		res.User.Encode(info.User, roles)
	}

	return res
}

type AuthResponse struct {
	Token string `json:"token"`
} //	@name	AuthResponse
