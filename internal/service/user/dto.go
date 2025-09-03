package user

import (
	"errors"
	"net/mail"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/auth_request"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/user_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_users"

	"github.com/google/uuid"
)

type CreateUserDTO struct {
	Email      string
	Password   string
	Location   string
	Language   string
	RateSource models.RateSource
	Role       models.UserRole
	Mnemonic   string
}

func (s CreateUserDTO) Validate() error {
	if s.Email == "" {
		return errors.New("email is required and cannot be empty")
	}

	if _, err := mail.ParseAddress(s.Email); err != nil {
		return errors.New("email must be a valid email address")
	}

	if s.Password == "" {
		return errors.New("password is required and cannot be empty")
	}

	if s.Location == "" {
		return errors.New("location is required and cannot be empty")
	}

	return nil
}

func (s CreateUserDTO) ToDbCreateParams() repo_users.CreateParams {
	return repo_users.CreateParams{
		Email:      s.Email,
		Password:   s.Password,
		Location:   s.Location,
		Language:   s.Language,
		RateSource: s.RateSource,
	}
}

func RequestToCreateUserDTO(req *auth_request.RegisterRequest) *CreateUserDTO {
	return &CreateUserDTO{
		Email:      req.Email,
		Password:   req.Password,
		Location:   req.Location,
		Language:   req.Language,
		RateSource: models.RateSourceBinance,
	}
}

type UpdateUserDTO struct {
	ID           uuid.UUID            `json:"id"`
	Location     string               `json:"location"`
	Language     string               `json:"language"`
	RateSource   models.RateSource    `json:"rate_source"`
	ExchangeSlug *models.ExchangeSlug `json:"exchange_slug"`
}

func (d *UpdateUserDTO) ToUpdateParams() repo_users.UpdateParams {
	return repo_users.UpdateParams{
		Location:   d.Location,
		Language:   d.Language,
		RateSource: d.RateSource,
	}
}

func RequestToUpdateUserDTO(req *user_request.UpdateRequest, id uuid.UUID) *UpdateUserDTO {
	return &UpdateUserDTO{
		ID:         id,
		Location:   req.Location,
		Language:   req.Language,
		RateSource: models.RateSource(req.RateSource),
	}
}
