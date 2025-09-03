package withdrawal_requests

import (
	"errors"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CreateAddressBookRequest struct {
	Address              string             `json:"address" validate:"required"`
	CurrencyID           string             `json:"currency_id" validate:"required_if=IsUniversal false IsEVM false"`
	IsUniversal          bool               `json:"is_universal"`
	IsEVM                bool               `json:"is_evm"`
	Name                 *string            `json:"name"`
	Tag                  *string            `json:"tag"`
	Blockchain           *models.Blockchain `json:"blockchain" validate:"required_if=IsUniversal true"`
	CreateWithdrawalRule *bool              `json:"create_withdrawal_rule"`
	TOTP                 string             `json:"totp" validate:"required,len=6"`
}

// Validate performs custom validation for CreateAddressBookRequest
func (r *CreateAddressBookRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Check mutual exclusivity
	if r.IsUniversal && r.IsEVM {
		return errors.New("only one type can be specified: is_universal or is_evm")
	}

	if !r.IsUniversal && !r.IsEVM {
		if r.CurrencyID == "" {
			return errors.New("currency_id is required for simple addresses")
		}
	}

	if r.IsUniversal {
		if r.Blockchain == nil {
			return errors.New("blockchain is required for universal addresses")
		}
		if err := r.Blockchain.Valid(); err != nil {
			return errors.New("invalid blockchain")
		}
	}

	return nil
}

type UpdateAddressBookRequest struct {
	Name *string `json:"name"`
	Tag  *string `json:"tag"`
	TOTP string  `json:"totp" validate:"required,len=6"`
}

type DeleteAddressBookRequest struct {
	DeleteWithdrawalRule *bool              `json:"delete_withdrawal_rule"`
	TOTP                 string             `json:"totp" validate:"required,len=6"`
	IsEVM                bool               `json:"is_evm"`
	IsUniversal          bool               `json:"is_universal"`
	ID                   *string            `json:"id" validate:"required_if=IsUniversal false IsEVM false"`
	Address              *string            `json:"address" validate:"required_with=IsEVM,required_with=IsUniversal"`
	Blockchain           *models.Blockchain `json:"blockchain" validate:"required_if=IsUniversal true"`
}

// Validate performs custom validation for DeleteAddressBookRequest
func (r *DeleteAddressBookRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Check mutual exclusivity
	if r.IsEVM && r.IsUniversal {
		return errors.New("only one type can be specified: is_evm or is_universal")
	}

	if !r.IsEVM && !r.IsUniversal {
		if r.ID == nil {
			return errors.New("id is required for individual address deletion")
		}

		_, err := uuid.Parse(*r.ID)
		if err != nil {
			return errors.New("invalid address ID format")
		}
	}
	if r.IsUniversal {
		if r.Blockchain == nil {
			return errors.New("blockchain is required for universal address deletion")
		}
		if err := r.Blockchain.Valid(); err != nil {
			return errors.New("invalid blockchain")
		}
	}
	if r.IsEVM {
		if r.Address == nil {
			return errors.New("address is required for EVM address deletion")
		}
	}

	return nil
}

type AddWithdrawalRuleRequest struct {
	TOTP        string             `json:"totp" validate:"required,len=6"`
	IsEVM       bool               `json:"is_evm"`
	IsUniversal bool               `json:"is_universal"`
	ID          *string            `json:"id" validate:"required_without_all=IsEVM IsUniversal"`
	Address     *string            `json:"address" validate:"required_with=IsEVM,required_with=IsUniversal"`
	Blockchain  *models.Blockchain `json:"blockchain" validate:"required_if=IsUniversal true"`
}

// Validate performs custom validation for AddWithdrawalRuleRequest
func (r *AddWithdrawalRuleRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Check mutual exclusivity
	if r.IsEVM && r.IsUniversal {
		return errors.New("only one type can be specified: is_evm or is_universal")
	}

	return nil
}
