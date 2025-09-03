package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/system_response"
	"github.com/dv-net/dv-merchant/internal/models"
)

func SystemInfoModelToResponse(info *models.SystemInfo) *system_response.SystemInfoResponse {
	return &system_response.SystemInfoResponse{
		AppProfile:        info.AppProfile,
		Initialized:       info.Initialized,
		RootUserExists:    info.RootUserExists,
		RegistrationState: info.RegistrationState,
		IsCaptchaEnabled:  info.IsTurnstileEnabled,
		SiteKey:           info.TurnstileSiteKey,
	}
}
