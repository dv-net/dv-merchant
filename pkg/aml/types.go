package aml

// ProviderSlug defines AML provider identity
type ProviderSlug string

const (
	ProviderSlugAMLBot ProviderSlug = "aml_bot"
	ProviderSlugBitOK  ProviderSlug = "bitok"
)

// AuthKeyType defines credential types
type AuthKeyType string

const (
	KeyAccessKeyID AuthKeyType = "access_key_id"
	KeyAccessID    AuthKeyType = "access_id"
	KeyAccessKey   AuthKeyType = "access_key"
	KeySecret      AuthKeyType = "secret"
)
