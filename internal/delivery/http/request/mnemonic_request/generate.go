package mnemonic_request

type GenerateMnemonicRequest struct {
	Length int `json:"length" validate:"oneof=12 24"`
}
