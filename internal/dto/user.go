package dto

import "github.com/dv-net/dv-merchant/internal/delivery/http/request/user_request"

type ChangePasswordDTO struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func RequestToChangePasswordDTO(req *user_request.ChangePasswordInternalRequest) ChangePasswordDTO {
	return ChangePasswordDTO{
		OldPassword: req.PasswordOld,
		NewPassword: req.PasswordNew,
	}
}
