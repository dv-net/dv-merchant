package webhook_response

type SendTestResult struct {
	ResponseStatus string `json:"response_status"`
	ResponseBody   string `json:"response_body"`
	RequestBody    string `json:"request_body"`
	ResponseCode   int    `json:"response_code"`
} //	@name	SendTestResult
