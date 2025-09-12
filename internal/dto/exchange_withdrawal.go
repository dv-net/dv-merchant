package dto

import "github.com/dv-net/dv-merchant/internal/delivery/http/request/exchange_request"

type CreateWithdrawalSettingDTO struct {
	Address    string `json:"address" validate:"required"`
	MinAmount  string `json:"min_amount" validate:"required"`
	CurrencyID string `json:"currency_id" validate:"required"`
	Chain      string `json:"chain" validate:"required"`
}

func RequestToCreateWithdrawalSettingDTO(req *exchange_request.CreateWithdrawalSettingRequest) *CreateWithdrawalSettingDTO {
	return &CreateWithdrawalSettingDTO{
		Address:    req.Address,
		MinAmount:  req.MinAmount,
		CurrencyID: req.CurrencyID,
		Chain:      req.Chain,
	}
}
