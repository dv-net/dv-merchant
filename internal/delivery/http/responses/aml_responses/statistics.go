package aml_responses

type StatisticsResponse struct {
	CheckedToday    int64 `json:"checked_today"`
	SuccessfulToday int64 `json:"successful_today"`
	FailedToday     int64 `json:"failed_today"`
} //	@name	AMLStatisticsResponse
