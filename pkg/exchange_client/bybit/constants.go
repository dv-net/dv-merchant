package bybit

const (
	MAINNET = "https://api.bybit.com"

	// Headers
	TimestampKey  = "X-BAPI-TIMESTAMP"
	SignatureKey  = "X-BAPI-SIGN"
	APIRequestKey = "X-BAPI-API-KEY"
	RecvWindowKey = "X-BAPI-RECV-WINDOW"
	SignTypeKey   = "X-BAPI-SIGN-TYPE"
)

var (
	ErrAPIKeyInvalid    int64 = 10003
	ErrAPISecretInvalid int64 = 10004
	ErrIPWhitelist      int64 = 10010
)
