package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/setting_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/settings_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/setting"
)

func FromSettingModelToResponse(model *models.Setting) *settings_response.SettingResponse {
	return &settings_response.SettingResponse{
		Name:  model.Name,
		Value: model.Value,
	}
}

func FromSettingModelToResponses(models ...*models.Setting) []*settings_response.SettingResponse {
	res := make([]*settings_response.SettingResponse, 0, len(models))
	for _, model := range models {
		res = append(res, FromSettingModelToResponse(model))
	}
	return res
}

func FromSettingRequestToDto(reqDto setting_request.Setting) setting.UpdateDto {
	return setting.UpdateDto{
		Name:  reqDto.Name,
		Value: reqDto.Value,
	}
}
