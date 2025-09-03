package models

type AMLSlug string

const (
	AMLSlugAMLBot AMLSlug = "aml_bot"
	AMLSlugBitOK  AMLSlug = "bit_ok"
)

func (s AMLSlug) String() string {
	return string(s)
}

func (s AMLSlug) Valid() bool {
	switch s {
	case AMLSlugAMLBot, AMLSlugBitOK:
		return true
	default:
		return false
	}
}
