package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/dvadmin_response"
	"github.com/dv-net/dv-merchant/internal/service/user"
)

func FromAdminOwnerDataToResponse(dto user.OwnerData) dvadmin_response.OwnerData {
	return dvadmin_response.OwnerData{
		IsAuthorized: dto.IsAuthorized,
		Balance:      dto.Balance,
		Telegram:     dto.Telegram,
		OwnerID:      dto.OwnerID,
	}
}
