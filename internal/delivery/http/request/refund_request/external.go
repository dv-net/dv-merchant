package refund_request

type ExternalRefundEstimateRequest struct {
	DestinationAddress string `json:"destination_address" validate:"required"`
} //	@name	ExternalRefundEstimateRequest

type ExternalRefundSendCodeRequest struct {
	DestinationAddress string `json:"destination_address" validate:"required"`
} //	@name	ExternalRefundSendCodeRequest

type ExternalRefundCreateRequest struct {
	DestinationAddress string `json:"destination_address" validate:"required"`
	// Code is the one-time code sent to the wallet email, required when the address is not a deposit sender
	Code string `json:"code" validate:"omitempty,len=6,alphanum"`
} //	@name	ExternalRefundCreateRequest
