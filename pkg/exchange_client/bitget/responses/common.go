package responses

import (
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/models"
)

type ResponseCode = int64

const (
	ResponseCodeOK                  ResponseCode = 00000
	ResponseCodeEmptyAccessKey      ResponseCode = 40001
	ResponseCodeInvalidAccessKey    ResponseCode = 40006
	ResponseCodeEmptyPassphrase     ResponseCode = 40011
	ResponseCodeInvalidPassphrase   ResponseCode = 40012
	ResponseCodeIncorrectPermission ResponseCode = 40014
	ResponseCodeInvalidIP           ResponseCode = 40018
	ResponseWithdrawalKYC           ResponseCode = 40027
	ResponseCodeInvalidAPIKey       ResponseCode = 40037
	ResponseCompleteKYC             ResponseCode = 40101
	ResponseCodeTemporaryDisabled   ResponseCode = 40127
	ResponseNotPerformedKYC         ResponseCode = 59005
)

type CommonResponse struct {
	Code        ResponseCode `json:"code,string"`
	Msg         string       `json:"msg"`
	RequestTime int64        `json:"request_time"`
}

type (
	AllAccountBalanceResponse struct {
		CommonResponse
		Data []*models.AccountBalance `json:"data,omitempty"`
	}
	ServerTimeResponse struct {
		CommonResponse
		Data *models.ServerTime `json:"data,omitempty"`
	}
)
