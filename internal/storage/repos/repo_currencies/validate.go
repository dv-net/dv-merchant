package repo_currencies

import "errors"

func (s CreateParams) Validate() error {
	if s.ID == "" {
		return errors.New("empty ID")
	}

	if s.Code == "" {
		return errors.New("empty Code")
	}

	if s.Name == "" {
		return errors.New("empty Name")
	}

	if err := s.Blockchain.Valid(); err != nil {
		return err
	}
	return nil
}
