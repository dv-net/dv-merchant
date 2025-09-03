package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/log_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/log"
)

func GetAllLogTypesResponse(m []*models.LogType) log_response.GetLogTypesResponse {
	r := log_response.GetLogTypesResponse{}
	for _, item := range m {
		v := log_response.LogTypeData{
			ID:        item.ID,
			Slug:      item.Slug,
			Title:     item.Title,
			CreatedAt: item.CreatedAt.Time,
		}
		r.Items = append(r.Items, v)
	}
	return r
}

func GetAllLogsResponse(m []*log.InfoDTO) log_response.GetLogsResponse {
	r := log_response.GetLogsResponse{}
	for _, item := range m {
		v := log_response.LogData{
			ProcessID: item.ProcessID,
			Failure:   item.Failure,
			CreatedAt: item.CreatedAt,
			Messages:  item.Messages,
		}
		r.Items = append(r.Items, v)
	}
	return r
}
