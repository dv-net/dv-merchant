package exchange

import (
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
)

type KeysExchangeDTO struct {
	Name  string  `json:"name"`
	Title string  `json:"title"`
	Value *string `json:"value"`
}

type WithExchangeKeysDTO struct {
	Exchange            string              `json:"exchange"`
	Slug                models.ExchangeSlug `json:"slug"`
	ExchangeConnectedAt time.Time           `json:"exchange_connected_at"`
	Keys                []KeysExchangeDTO   `json:"keys"`
}

type ActiveExchangeListDTO struct {
	Exchanges       []WithExchangeKeysDTO `json:"exchanges"`
	CurrentExchange *string               `json:"current_exchange"`
	SwapState       *string               `json:"swap_state"`
	WithdrawalState *string               `json:"withdrawal_state"`
}

type UserExchangeOrderModel struct {
	Exchange        string    `json:"exchange" csv:"exchange_name" excel:"exchange_name"`
	ExchangeID      string    `json:"exchange_id" csv:"exchange_id" excel:"exchange_id"`
	ExchangeOrderID string    `json:"exchange_order_id" csv:"exchange_order_id" excel:"exchange_order_id"`
	ClientOrderID   string    `json:"client_order_id" csv:"client_order_id" excel:"client_order_id"`
	Symbol          string    `json:"symbol" csv:"order_symbol" excel:"order_symbol"`
	Side            string    `json:"side" csv:"order_side" excel:"order_side"`
	Amount          string    `json:"amount" csv:"order_amount" excel:"order_amount"`
	AmountUsd       string    `json:"amount_usd" csv:"order_amount_usd" excel:"order_amount_usd"`
	FailReason      string    `json:"fail_reason" csv:"order_fail_reason" excel:"order_fail_reason"`
	Status          string    `json:"status" csv:"order_status" excel:"order_status"`
	OrderCreatedAt  time.Time `json:"order_created_at" csv:"order_created_at" excel:"order_created_at"`
}
