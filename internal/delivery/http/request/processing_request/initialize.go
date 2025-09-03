package processing_request

type InitializeRequest struct {
	// Host (http://domain.example:574 or https://127.0.0.1:999)
	Host                  string `json:"host" validate:"required,url"`
	CallbackDomain        string `json:"callback_domain" validate:"required"`
	MerchantDomain        string `json:"merchant_domain" validate:"required"`
	MerchantPayFormDomain string `json:"merchant_pay_form_domain" validate:"required"`
} // @name InitializeProcessingRequest
