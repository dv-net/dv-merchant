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

func (s AMLSlug) Label() string {
	switch s {
	case AMLSlugAMLBot:
		return "AML Bot"
	case AMLSlugBitOK:
		return "BitOK"
	case AMLSlugCoinKyt:
		return "Coin KYT"
	default:
		return string(s)
	}
}

func (s AMLSlug) Valid() bool {
	switch s {
	case AMLSlugAMLBot, AMLSlugBitOK, AMLSlugCoinKyt:
		return true
	default:
		return false
	}
}

type AMLProvider struct {
	Slug  AMLSlug `json:"slug"`
	Label string  `json:"label"`
}
