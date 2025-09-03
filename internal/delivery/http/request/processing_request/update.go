package processing_request

type UpdateProcessingCallbackDomain struct {
	Domain string `json:"domain" validate:"required,min=1,max=255"`
}
