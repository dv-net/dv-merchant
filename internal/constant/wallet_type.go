package constant

type WalletAddressType string

const (
	WalletAddress WalletAddressType = "wallet"
	RotateAddress WalletAddressType = "rotate"
)

func (t WalletAddressType) String() string {
	return string(t)
}
