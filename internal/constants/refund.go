package constants

const (
	RefundStatusPendingReview = "pending_review"
	RefundStatusRejected      = "rejected"
	RefundStatusProcessing    = "processing"
	RefundStatusScheduled     = "scheduled"
	RefundStatusCanceled      = "canceled"
	RefundStatusCompleted     = "completed"
	RefundStatusFailed        = "failed"
)

const (
	RefundAuthMethodEmail = "email"
	RefundAuthMethodNone  = "none"
	// RefundAuthMethodAPI is a refund to the deposit sender requested by the store through the external api
	RefundAuthMethodAPI = "api"
	// RefundAuthMethodAPIEmail is a refund requested by the store through the external api,
	// confirmed by the code sent to the wallet email
	RefundAuthMethodAPIEmail = "api_email"
)
