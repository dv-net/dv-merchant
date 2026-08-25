package aml_responses

import "github.com/dv-net/dv-merchant/internal/service/aml"

type StatisticsResponse struct {
	CheckedToday    int64 `json:"checked_today"`
	SuccessfulToday int64 `json:"successful_today"`
	FailedToday     int64 `json:"failed_today"`
} //	@name	AMLStatisticsResponse

func NewStatisticsResponse(dto *aml.StatisticsDTO) *StatisticsResponse {
	return &StatisticsResponse{
		CheckedToday:    dto.CheckedToday,
		SuccessfulToday: dto.SuccessfulToday,
		FailedToday:     dto.FailedToday,
	}
}
