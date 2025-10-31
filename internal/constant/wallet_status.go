package constant

type WalletStatus string

const (
	WalletStatusAvailable WalletStatus = "available" // The address is free, can be given out
	WalletStatusReserved  WalletStatus = "reservers" // Reserved for pickup
	WalletStatusLocked    WalletStatus = "locked"    // locked address
	WalletStatusStatic    WalletStatus = "static"    // used for static wallet
)

func (w WalletStatus) String() string {
	return string(w)
}
