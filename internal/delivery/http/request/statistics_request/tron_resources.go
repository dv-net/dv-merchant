package statistics_request

type FetchTronStatsRequest struct {
	Resolution string  `json:"resolution" query:"resolution" validate:"required,oneof=day hour"`
	DateFrom   *string `json:"date_from" query:"date_from"`
	DateTo     *string `json:"date_to" query:"date_to"`
} // @name FetchTronStatsRequest
