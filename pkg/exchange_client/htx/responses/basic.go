//nolint:tagliatelle
package responses

type Status string

func (o Status) String() string { return string(o) }

const (
	StatusOK    Status = "ok"
	StatusError Status = "error"
)

type Basic struct {
	Status  Status `json:"status,omitempty"`
	Code    any    `json:"code,omitempty"`
	Msg     string `json:"msg,omitempty"`
	ErrCode any    `json:"err-code,omitempty"`
	ErrMsg  string `json:"err-msg,omitempty"`
	Message string `json:"message,omitempty"` // user api
	OK      any    `json:"ok,omitempty"`      // user api
}

type AdditionalData struct {
	Timestamp int64 `json:"timestamp,omitempty"`
	Full      int   `json:"full,omitempty"`
}
