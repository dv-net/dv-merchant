package transaction_response

import (
	"time"

	"github.com/shopspring/decimal"
)

type LoadUserTransactionsResponse struct{} // @name LoadUserTransactionsResponse

type SearchUserTransactionResponse struct{} // @name SearchUserTransactionResponse

type TransactionResponse struct {
	ID                 string              `json:"id" format:"uuid"`
	UserID             string              `json:"user_id" format:"uuid"`
	StoreID            string              `json:"store_id" format:"uuid"`
	ReceiptID          string              `json:"receipt_id" format:"uuid"`
	WalletID           string              `json:"wallet_id" format:"uuid"`
	CurrencyID         string              `json:"currency_id"`
	Blockchain         string              `json:"blockchain"`
	TxHash             string              `json:"tx_hash"`
	BcUniqKey          *string             `json:"bc_uniq_key"`
	Type               string              `json:"type"`
	FromAddress        string              `json:"from_address"`
	ToAddress          string              `json:"to_address"`
	Amount             decimal.Decimal     `json:"amount"`
	AmountUsd          decimal.NullDecimal `json:"amount_usd"`
	Fee                decimal.Decimal     `json:"fee"`
	WithdrawalIsManual bool                `json:"withdrawal_is_manual"`
	NetworkCreatedAt   time.Time           `json:"network_created_at" format:"date-time"`
	CreatedAt          time.Time           `json:"created_at" format:"date-time"`
	UpdatedAt          time.Time           `json:"updated_at" format:"date-time"`
	CreatedAtIndex     int                 `json:"created_at_index"`
} // @name TransactionResponse

type ShortTransactionInfoListResponse struct {
	Confirmed   []ShortTransactionResponse `json:"confirmed"`
	Unconfirmed []ShortTransactionResponse `json:"unconfirmed"`
} // @name ShortTransactionInfoListResponse

type ShortTransactionResponse struct {
	CurrencyCode string    `json:"currency_code"`
	Hash         string    `json:"hash"`
	Amount       string    `json:"amount"`
	AmountUSD    string    `json:"amount_usd"`
	Type         string    `json:"type"`
	CreatedAt    time.Time `json:"created_at"`
} // @name ShortTransactionResponse
