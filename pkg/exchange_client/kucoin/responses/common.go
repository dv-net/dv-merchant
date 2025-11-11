package responses

type ErrorCode = int

const (
	SuccessCodeOK int = 200000
)

const (
	ErrorCodeBalanceInsufficient ErrorCode = 200004
	ErrorCodeMissingAPICreds     ErrorCode = 400001
	ErrorCodeInvalidTimestamp    ErrorCode = 400002
	ErrorCodeInvalidAPIKey       ErrorCode = 400003
	ErrorCodeInvalidPassphrase   ErrorCode = 400004
	ErrorCodeInvalidSignature    ErrorCode = 400005
	ErrorCodeIPWhitelist         ErrorCode = 400006
	ErrorCodeMinOrderValue       ErrorCode = 400760 // Order value below minimum (e.g., "should more then 0.1 USDT")
	ErrorWithdrawalTooFast       ErrorCode = 115004
	// Rate limit errors
	ErrorCodeRateLimitExceeded ErrorCode = 429000 // Both system-level and user-level rate limits
)

type Basic struct {
	Code int `json:"code,string"`
	Msg  any `json:"msg,omitempty"`
}
