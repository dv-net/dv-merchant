package models

type AMLSlug string

const (
	AMLSlugAMLBot  AMLSlug = "aml_bot"
	AMLSlugBitOK   AMLSlug = "bit_ok"
	AMLSlugCoinKyt AMLSlug = "coin_kyt"
)

func (s AMLSlug) String() string {
	return string(s)
}

func (s AMLSlug) Valid() bool {
	switch s {
	case AMLSlugAMLBot, AMLSlugBitOK, AMLSlugCoinKyt:
		return true
	default:
		return false
	}
}
