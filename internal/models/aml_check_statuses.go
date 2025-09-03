package models

type AMLCheckStatus string

const (
	AmlCheckStatusPending AMLCheckStatus = "pending"
	AmlCheckStatusSuccess AMLCheckStatus = "success"
	AmlCheckStatusFailed  AMLCheckStatus = "failed"
)

func (s AMLCheckStatus) String() string {
	return string(s)
}
