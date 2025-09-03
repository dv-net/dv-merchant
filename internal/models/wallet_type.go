package models

type WalletType string

const (
	WalletTypeCold       WalletType = "cold"
	WalletTypeHot        WalletType = "hot"
	WalletTypeProcessing WalletType = "processing"
)

func (t *WalletType) String() string {
	if t == nil {
		return ""
	}

	return string(*t)
}
