package models

type CurrencyShort struct {
	ID            string      `db:"id" json:"id"`
	Code          string      `db:"code" json:"code"`
	Precision     int16       `db:"precision" json:"precision"`
	Name          string      `db:"name" json:"name"`
	Blockchain    *Blockchain `db:"blockchain" json:"blockchain"`
	IsBitcoinLike bool        `db:"is_bitcoin_like" json:"is_bitcoin_like"`
	IsEVMLike     bool        `db:"is_evm_like" json:"is_evm_like"`
} // @name CurrencyShort
