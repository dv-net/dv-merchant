package models

import "github.com/google/uuid"

type SettingModelName string

const (
	SettingModelNameUser SettingModelName = "User"
)

func (u *User) ModelID() uuid.NullUUID {
	return uuid.NullUUID{UUID: u.ID, Valid: true}
}

func (u *User) ModelName() *string {
	res := string(SettingModelNameUser)
	return &res
}
