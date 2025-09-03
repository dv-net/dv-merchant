package converters

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/responses/withdrawal_response"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/withdrawal_wallet"

	"github.com/google/uuid"
)

func FromWithdrawalWalletAddressToResponse(o *models.WithdrawalWalletAddress) *withdrawal_response.WithdrawalWalletAddressResponse {
	w := &withdrawal_response.WithdrawalWalletAddressResponse{
		ID:                 o.ID,
		WithdrawalWalletID: o.WithdrawalWalletID,
		Name:               o.Name,
		Address:            o.Address,
		CreatedAt:          o.CreatedAt.Time,
		UpdatedAt:          o.UpdatedAt.Time,
		DeletedAt:          o.DeletedAt.Time,
	}
	return w
}

func FromWithdrawalWalletAddressToResponses(o ...*models.WithdrawalWalletAddress) []*withdrawal_response.WithdrawalWalletAddressResponse {
	w := make([]*withdrawal_response.WithdrawalWalletAddressResponse, 0, len(o))
	for _, v := range o {
		w = append(w, FromWithdrawalWalletAddressToResponse(v))
	}
	return w
}

func FromWithdrawalWithAddressToResponse(o *withdrawal_wallet.WithdrawalWithAddress) *withdrawal_response.WithdrawalWithAddressResponse {
	return &withdrawal_response.WithdrawalWithAddressResponse{
		ID:              o.ID,
		Status:          o.Status,
		NativeBalance:   o.MinBalanceNative,
		USDBalance:      o.MinBalanceUsd,
		Addressees:      FromWithdrawalWalletAddressToResponses(o.Addressees...),
		Interval:        o.Interval,
		Currency:        o.Currency,
		Rate:            o.Rate,
		LowBalanceRules: FromMultiWithdrawalRuleToLowBalanceRulesResponse(o.MultiWithdrawal),
	}
}

func FromWithdrawalWithAddressToResponses(o ...*withdrawal_wallet.WithdrawalWithAddress) []*withdrawal_response.WithdrawalWithAddressResponse {
	w := make([]*withdrawal_response.WithdrawalWithAddressResponse, 0, len(o))
	for _, v := range o {
		w = append(w, FromWithdrawalWithAddressToResponse(v))
	}
	return w
}

func FromWithdrawalWithAddressToByCurrencyResponse(dto *withdrawal_wallet.WithdrawalWithAddress) *withdrawal_response.WithdrawalRulesByCurrencyResponse {
	return &withdrawal_response.WithdrawalRulesByCurrencyResponse{
		ID:              dto.ID,
		Status:          dto.Status,
		NativeBalance:   dto.MinBalanceNative,
		USDBalance:      dto.MinBalanceUsd,
		Addressees:      dto.Addressees,
		Interval:        dto.Interval,
		Currency:        dto.Currency,
		Rate:            dto.Rate,
		LowBalanceRules: FromMultiWithdrawalRuleToLowBalanceRulesResponse(dto.MultiWithdrawal),
	}
}

func FromMultiWithdrawalRuleToLowBalanceRulesResponse(dto withdrawal_wallet.MultiWithdrawalRuleDTO) withdrawal_response.LowBalanceWithdrawalRuleResponse {
	return withdrawal_response.LowBalanceWithdrawalRuleResponse{
		Mode:          dto.Mode,
		ManualAddress: dto.ManualAddress,
	}
}

func FromProcessingWithdrawalToResponse(model models.WithdrawalFromProcessingWallet) withdrawal_response.ProcessingWithdrawalResponse {
	var transferID *uuid.UUID
	if model.TransferID.Valid {
		transferID = &model.TransferID.UUID
	}

	return withdrawal_response.ProcessingWithdrawalResponse{
		ID:          model.ID,
		TransferID:  transferID,
		StoreID:     model.StoreID,
		CurrencyID:  model.CurrencyID,
		AddressFrom: model.AddressFrom,
		AddressTo:   model.AddressTo,
		Amount:      model.Amount.String(),
		AmountUsd:   model.AmountUsd.String(),
		CreatedAt:   model.CreatedAt.Time,
	}
}
