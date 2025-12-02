package withdrawal_response

import (
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

type AddressBookEntryResponse struct {
	ID                   uuid.UUID         `json:"id"`
	Address              string            `json:"address,omitempty"`
	CurrencyID           string            `json:"currency_id"`
	Name                 *string           `json:"name"`
	Tag                  *string           `json:"tag"`
	Blockchain           models.Blockchain `json:"blockchain,omitempty"`
	SubmittedAt          string            `json:"submitted_at"`
	WithdrawalRuleExists bool              `json:"withdrawal_rule_exists"`
} //	@name	AddressBookEntryResponse

type AddressBookEntryResponseShort struct {
	ID                   uuid.UUID `json:"id"`
	CurrencyID           string    `json:"currency_id"`
	WithdrawalRuleExists bool      `json:"withdrawal_rule_exists"`
} //	@name	AddressBookEntryResponseShort

type UniversalAddressGroupResponse struct {
	Address       string                           `json:"address"`
	Name          *string                          `json:"name"`
	Tag           *string                          `json:"tag"`
	Blockchain    models.Blockchain                `json:"blockchain"`
	IsUniversal   bool                             `json:"is_universal"`
	Currencies    []*AddressBookEntryResponseShort `json:"currencies"`
	SubmittedAt   string                           `json:"submitted_at"`
	CurrencyCount int                              `json:"currency_count"`
} //	@name	UniversalAddressGroupResponse

type EVMAddressGroupResponse struct {
	Address       string                           `json:"address"`
	Name          *string                          `json:"name"`
	Tag           *string                          `json:"tag"`
	IsEVM         bool                             `json:"is_evm"`
	Blockchains   []models.Blockchain              `json:"blockchains"`
	Currencies    []*AddressBookEntryResponseShort `json:"currencies"`
	SubmittedAt   string                           `json:"submitted_at"`
	CurrencyCount int                              `json:"currency_count"`
} //	@name	EVMAddressGroupResponse

type AddressBookListResponse struct {
	Addresses       []*AddressBookEntryResponse      `json:"addresses"`
	UniversalGroups []*UniversalAddressGroupResponse `json:"universal_groups"`
	EVMGroups       []*EVMAddressGroupResponse       `json:"evm_groups"`
} //	@name	AddressBookListResponse
