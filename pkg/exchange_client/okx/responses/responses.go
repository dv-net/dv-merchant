package responses

type (
	Basic struct {
		Code any    `json:"code"`
		Msg  string `json:"msg,omitempty"`
	}
)

type ErrorCode = int

const (
	ErrorCodeInvalidAccessKey       ErrorCode = 50103
	ErrorCodeInvalidPassphraseEmpty ErrorCode = 50104
	ErrorCodeInvalidPassphrase      ErrorCode = 50105
	ErrorCodeIPWhitelist            ErrorCode = 50110
	ErrorCodeInvalidSignature       ErrorCode = 50113
	ErrorCodeAPIKeyNotExists        ErrorCode = 50119
	ErrorCodeInvalidTimestamp       ErrorCode = 50107
)

// 	50103	401	Request header "OK-ACCESS-KEY" cannot be empty.
// 50104	401	Request header "OK-ACCESS-PASSPHRASE" cannot be empty.
// 50105	401	Request header "OK-ACCESS-PASSPHRASE" incorrect.
// 50106	401	Request header "OK-ACCESS-SIGN" cannot be empty.
// 50107	401	Request header "OK-ACCESS-TIMESTAMP" cannot be empty.
