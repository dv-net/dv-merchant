package repo_users

import (
	"errors"
	"net/mail"
)

func (s CreateParams) Validate() error {
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
