package transaction_response

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionInfoResponse struct {
	ID                 uuid.UUID               `json:"id"`
	IsConfirmed        bool                    `json:"is_confirmed"`
	UserID             uuid.UUID               `json:"user_id"`
	StoreID            *uuid.UUID              `json:"store_id"`
	ReceiptID          *uuid.UUID              `json:"receipt_id"`
	Wallet             WalletInfoResponse      `json:"wallet"`
	CurrencyID         string                  `json:"currency_id"`
	Blockchain         string                  `json:"blockchain"`
	TxHash             string                  `json:"tx_hash"`
	BcUniqKey          string                  `json:"bc_uniq_key"`
	Type               string                  `json:"type"`
	FromAddress        string                  `json:"from_address"`
	ToAddress          string                  `json:"to_address"`
	Amount             decimal.Decimal         `json:"amount"`
	AmountUsd          *decimal.Decimal        `json:"amount_usd"`
	Fee                decimal.Decimal         `json:"fee"`
	WithdrawalIsManual bool                    `json:"withdrawal_is_manual"`
	NetworkCreatedAt   *time.Time              `json:"network_created_at"`
	WebhookHistory     []WhSendHistoryResponse `json:"webhook_history"`
	CreatedAt          *time.Time              `json:"created_at"`
	UpdatedAt          *time.Time              `json:"updated_at"`
} // @name TransactionInfoResponse

type WhSendHistoryResponse struct {
	ID         uuid.UUID  `json:"id"`
	StoreID    uuid.UUID  `json:"store_id"`
	WhType     string     `json:"wh_type"`
	URL        string     `json:"url"`
	IsSuccess  bool       `json:"is_success"`
	Request    string     `json:"request"`
	Response   *string    `json:"response"`
	StatusCode int        `json:"status_code"`
	CreatedAt  *time.Time `json:"created_at"`
} // @name WhSendHistoryResponse

type WalletInfoResponse struct {
	ID              uuid.UUID  `json:"id"`
	WalletStoreID   uuid.UUID  `json:"wallet_store_id"`
	StoreExternalID string     `json:"store_external_id"`
	WalletCreatedAt time.Time  `json:"wallet_created_at"`
	WalletUpdatedAt *time.Time `json:"wallet_updated_at"`
} // @name WalletInfoResponse
