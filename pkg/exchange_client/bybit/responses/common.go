//nolint:tagliatelle
package responses

type BaseResponse[T any] struct {
	RetCode    int      `json:"retCode"`
	RetMsg     string   `json:"retMsg"`
	Result     T        `json:"result"`
	RetExtInfo struct{} `json:"retExtInfo"`
	Time       int64    `json:"time"`
}
