//nolint:tagliatelle
package responses

import (
	htxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/models"
)

// Account client responses
type (
	GetAccounts struct {
		Basic
		Accounts []*htxmodels.Account `json:"data,omitempty"`
	}
	GetAccountBalance struct {
		Basic
		Balances *htxmodels.AccountBalance `json:"data,omitempty"`
	}
)
