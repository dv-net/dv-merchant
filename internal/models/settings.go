package models

import "github.com/google/uuid"

type SettingModelName string

const (
	SettingModelNameUser  SettingModelName = "User"
	SettingModelNameStore SettingModelName = "Store"
)

func (u *User) ModelID() uuid.NullUUID {
	return uuid.NullUUID{UUID: u.ID, Valid: true}
}

func (u *User) ModelName() *string {
	return new(string(SettingModelNameUser))
}

func (s *Store) ModelID() uuid.NullUUID {
	return uuid.NullUUID{UUID: s.ID, Valid: true}
}

func (s *Store) ModelName() *string {
	return new(string(SettingModelNameStore))
}
