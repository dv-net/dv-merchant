package constants

type StoreVerificationStatus string

const (
	StoreVerificationStatusPending  StoreVerificationStatus = "PENDING"
	StoreVerificationStatusSuccess  StoreVerificationStatus = "SUCCESS"
	StoreVerificationStatusRejected StoreVerificationStatus = "REJECTED"
)

func (s StoreVerificationStatus) String() string {
	return string(s)
}
