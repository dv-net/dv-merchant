package withdrawal_response

import (
	"time"

	"github.com/google/uuid"
)

type ProcessingWithdrawalResponse struct {
	ID          uuid.UUID  `json:"id"`
	TransferID  *uuid.UUID `json:"transfer_id"`
	StoreID     uuid.UUID  `json:"store_id"`
	CurrencyID  string     `json:"currency_id"`
	AddressFrom string     `json:"address_from"`
	AddressTo   string     `json:"address_to"`
	Amount      string     `json:"amount"`
	AmountUsd   string     `json:"amount_usd"`
	CreatedAt   time.Time  `json:"created_at" format:"date-time"`
} //	@name	ProcessingWithdrawalResponse
