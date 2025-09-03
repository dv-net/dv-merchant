package wallet_request

import "github.com/dv-net/dv-merchant/internal/models"

type AddressConverterRequest struct {
	Address    string            `json:"address" validate:"required,min=16,max=255"`
	Blockchain models.Blockchain `json:"blockchain" validate:"required"`
}
