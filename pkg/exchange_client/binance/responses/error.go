package responses

type ErrorResponse struct {
	Code ResponseCode `json:"code"`
	Msg  string       `json:"msg"`
}

type ResponseCode = int64

const (
	ResponseCodeInvalidPermissionsOrIP ResponseCode = -2015
	ResponseCodeInvalidAPIKey          ResponseCode = -2014
	ResponseCodeInvalidSecretKey       ResponseCode = -1022
	ResponseCodeInvalidPermissionOnAPI ResponseCode = -1002
)
