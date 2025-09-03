package models

type CurrencyType string

const (
	CurrencyTypeFiat   CurrencyType = "fiat"
	CurrencyTypeCrypto CurrencyType = "crypto"
)

func (o CurrencyType) Valid() bool {
	switch o {
	case CurrencyTypeFiat, CurrencyTypeCrypto:
		return true
	}
	return false
}

func (o CurrencyType) String() string { return string(o) }
