package processing_request

type MnemonicRequest struct {
	Mnemonic string `json:"mnemonic" validate:"required"`
}
