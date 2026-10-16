package withdrawal_response

import (
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ExternalWithdrawalResponse struct {
	TransferID  uuid.UUID             `json:"transfer_id" format:"uuid"`
	Status      models.TransferStatus `json:"status"`
	FromAddress string                `json:"from_address"`
	AddressTo   string                `json:"address_to"`
	Amount      decimal.Decimal       `json:"amount"`
	AmountUsd   decimal.Decimal       `json:"amount_usd"`
	CurrencyID  string                `json:"currency_id"`
	Message     *string               `json:"message,omitempty"`
} //	@name	ExternalWithdrawalResponse

func NewExternalWithdrawalResponse(transfer *models.Transfer) ExternalWithdrawalResponse {
	res := ExternalWithdrawalResponse{
		TransferID: transfer.ID,
		Status:     transfer.Status,
		Amount:     transfer.Amount,
		AmountUsd:  transfer.AmountUsd,
		CurrencyID: transfer.CurrencyID,
		Message:    transfer.Message,
	}

	if len(transfer.FromAddresses) > 0 {
		res.FromAddress = transfer.FromAddresses[0]
	}

	if len(transfer.ToAddresses) > 0 {
		res.AddressTo = transfer.ToAddresses[0]
	}

	return res
}
