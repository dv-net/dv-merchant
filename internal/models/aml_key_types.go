package models

type AmlKeyType string

const (
	AmlKeyTypeAccessKeyID AmlKeyType = "access_key_id"
	AmlKeyTypeSecret      AmlKeyType = "secret_key"
	AmlKeyTypeAccessKey   AmlKeyType = "access_key"
	AmlKeyTypeAccessID    AmlKeyType = "access_id"
)
