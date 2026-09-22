package models

import (
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

// AddressRiskFlag is one entry of the wallet_addresses.risk_flags JSONB array,
// written by aml.Service.ApplyVerdict via wallet.Service.MarkAddressFlags. Slug is the
// risk_type of the "accept_and_flag" rule that fired — there is no separate canonical
// flag enum, the rule's own risk_type is the tag.
type AddressRiskFlag struct {
	Slug       string    `json:"slug"`
	AmlCheckID uuid.UUID `json:"aml_check_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// ParseAddressRiskFlags decodes a wallet_addresses.risk_flags column. A nil or empty
// payload, including the column's default literal "[]", decodes to a nil slice.
func ParseAddressRiskFlags(raw []byte) ([]AddressRiskFlag, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var flags []AddressRiskFlag
	if err := json.Unmarshal(raw, &flags); err != nil {
		return nil, err
	}

	return flags, nil
}
