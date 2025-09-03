//nolint:tagliatelle
package models

type APIKeyInformation struct {
	AccessKey   string `json:"accessKey"`
	Note        string `json:"note"`
	Permission  string `json:"permission"`
	IPAddresses string `json:"ipAddresses"`
	ValidDays   int    `json:"validDays"`
	Status      string `json:"status"`
	CreateTime  int64  `json:"createTime"`
	UpdateTime  int64  `json:"updateTime"`
}
