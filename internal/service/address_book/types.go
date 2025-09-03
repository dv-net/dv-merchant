package address_book

import (
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

type AddressBookEntry struct {
	ID          uuid.UUID          `json:"id"`
	Address     string             `json:"address"`
	CurrencyID  string             `json:"currency_id"`
	Name        *string            `json:"name"`
	Tag         *string            `json:"tag"`
	Blockchain  *models.Blockchain `json:"blockchain"`
	SubmittedAt string             `json:"submitted_at"`
}

type CreateAddressBookRequest struct {
	Address    string             `json:"address" validate:"required"`
	CurrencyID string             `json:"currency_id" validate:"required"`
	Name       *string            `json:"name"`
	Tag        *string            `json:"tag"`
	Blockchain *models.Blockchain `json:"blockchain"`
}

type UpdateAddressBookRequest struct {
	Name *string `json:"name"`
	Tag  *string `json:"tag"`
}

type AddressBookListResponse struct {
	Addresses []*AddressBookEntry `json:"addresses"`
}

type AddressType string

func (o AddressType) String() string { return string(o) }

const (
	AddressTypeUniversalAddress AddressType = "universal"
	AddressTypeEVMAddress       AddressType = "evm"
	AddressTypeSimpleAddress    AddressType = "simple"
)

type DeleteAddressDTO struct {
	UserID               uuid.UUID          `json:"user_id"`
	ID                   *uuid.UUID         `json:"id"`
	Address              *string            `json:"address"`
	Blockchain           *models.Blockchain `json:"blockchain"`
	DeleteWithdrawalRule bool               `json:"delete_withdrawal_rule"`
	IsEVM                bool               `json:"is_evm"`
	IsUniversal          bool               `json:"is_universal"`
}

type AddWithdrawalRuleDTO struct {
	UserID      uuid.UUID          `json:"user_id"`
	ID          *uuid.UUID         `json:"id"`
	Address     *string            `json:"address"`
	Blockchain  *models.Blockchain `json:"blockchain"`
	IsEVM       bool               `json:"is_evm"`
	IsUniversal bool               `json:"is_universal"`
	TOTP        string             `json:"totp"`
}
