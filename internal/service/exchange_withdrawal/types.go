package exchange_withdrawal

import (
	"time"

	"github.com/google/uuid"
)

type UpdateWithdrawalSettingDto struct {
	ID        uuid.UUID
	IsEnabled bool
}

type UserWithdrawalHistoryModel struct {
	ExchangeName    string    `json:"exchange_name" csv:"exchange_name" excel:"exchange_name"`
	ExchangeOrderID string    `json:"exchange_order_id" csv:"exchange_order_id" excel:"exchange_order_id"`
	ExchangeID      string    `json:"exchange_id" csv:"exchange_id" excel:"exchange_id"`
	ExchangeSlug    string    `json:"exchange_slug" csv:"exchange_slug" excel:"exchange_slug"`
	Address         string    `json:"address" csv:"address" excel:"address"`
	NativeAmount    string    `json:"native_amount" csv:"native_amount" excel:"native_amount"`
	FiatAmount      string    `json:"fiat_amount" csv:"fiat_amount" excel:"fiat_amount"`
	Currency        string    `json:"currency" csv:"currency" excel:"currency"`
	Chain           string    `json:"chain" csv:"chain" excel:"chain"`
	Status          string    `json:"status" csv:"status" excel:"status"`
	Txid            string    `json:"txid" csv:"txid" excel:"txid"`
	CreatedAt       time.Time `json:"created_at" csv:"created_at" excel:"created_at"`
	FailReason      string    `json:"fail_reason" csv:"fail_reason" excel:"fail_reason"`
}
