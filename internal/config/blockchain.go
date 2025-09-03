package config

type Blockchain struct {
	Bitcoin  BitcoinBlockchain
	Litecoin LitecoinBlockchain
	Dogecoin DogecoinBlockchain
}

type BitcoinBlockchain struct {
	AddressType string `json:"address_type" yaml:"address_type" usage:"change default generate address" default:"P2WPKH" example:"P2PKH / P2SH / P2WPKH / P2TR"`
}
type LitecoinBlockchain struct {
	AddressType string `json:"address_type" yaml:"address_type" usage:"change default generate address" default:"P2WPKH" example:"P2PKH / P2SH / P2WPKH / P2TR"`
}

type DogecoinBlockchain struct {
	AddressType string `json:"address_type" yaml:"address_type" usage:"change default generate address" default:"P2PKH" example:"P2PKH"`
}
