package exchange_request

import (
	"github.com/dv-net/dv-merchant/internal/models"
)

type KeyData struct {
	Name  string `json:"name" validate:"required,min=1,max=255"`
	Value string `json:"value" validate:"required,min=1,max=255"`
} //	@name	ExchangeUpdateKeyData

type UpdateKeys struct {
	Keys []KeyData `json:"keys" required:"true" validate:"required,dive"`
} //	@name	ExchangeUpdateKeysRequest

func (uk *UpdateKeys) ToMap() map[models.ExchangeKeyName]*string {
	res := make(map[models.ExchangeKeyName]*string, len(uk.Keys))
	for _, k := range uk.Keys {
		exKeyName := models.ExchangeKeyName(k.Name)
		if exKeyName.Valid() && k.Value != "" {
			res[exKeyName] = &k.Value
		}
	}
	return res
}
