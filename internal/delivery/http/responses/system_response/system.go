package system_response

import "github.com/dv-net/dv-merchant/internal/models"

type SystemInfoResponse struct {
	AppProfile        models.AppProfile `json:"app_profile"`
	Initialized       bool              `json:"initialized"`
	RootUserExists    bool              `json:"root_user_exists"`
	RegistrationState string            `json:"registration_state"`
	IsCaptchaEnabled  bool              `json:"is_captcha_enabled"`
	SiteKey           string            `json:"site_key"`
} // @name SystemInfoResponse
