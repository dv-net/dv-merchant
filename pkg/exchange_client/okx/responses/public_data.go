//nolint:tagliatelle
package responses

import (
	okxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/models"
)

type (
	GetSystemTime struct {
		Basic
		SystemTimes []*okxmodels.SystemTime `json:"data,omitempty"`
	}
	GetInstruments struct {
		Basic
		Instruments []*okxmodels.Instrument `json:"data,omitempty"`
	}
)
