package models

type ExchangeKeyName string

const (
	ExchangeKeyNameAccessKey  ExchangeKeyName = "access_key"
	ExchangeKeyNameSecretKey  ExchangeKeyName = "secret_key"
	ExchangeKeyNamePassPhrase ExchangeKeyName = "pass_phrase"
	ExchangeKeyNameAPIKey     ExchangeKeyName = "api_key"
)

func (ekn ExchangeKeyName) Valid() bool {
	switch ekn {
	case ExchangeKeyNameAccessKey, ExchangeKeyNameSecretKey, ExchangeKeyNamePassPhrase, ExchangeKeyNameAPIKey:
		return true
	default:
		return false
	}
}

func (ekn ExchangeKeyName) String() string { return string(ekn) }
