package constants

type StoreVerificationStatus string

const (
	StoreVerificationStatusPending            StoreVerificationStatus = "PENDING"
	StoreVerificationStatusSuccess            StoreVerificationStatus = "SUCCESS"
	StoreVerificationStatusRejected           StoreVerificationStatus = "REJECTED"
	StoreVerificationStatusNeedsClarification StoreVerificationStatus = "NEEDS_CLARIFICATION"
)

func (s StoreVerificationStatus) String() string {
	return string(s)
}
