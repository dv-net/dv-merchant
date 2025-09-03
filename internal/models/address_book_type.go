package models

type AddressBookType string

func (o AddressBookType) String() string { return string(o) }

const (
	AddressBookTypeSimple    AddressBookType = "simple"
	AddressBookTypeUniversal AddressBookType = "universal"
	AddressBookTypeEVM       AddressBookType = "evm"
)
