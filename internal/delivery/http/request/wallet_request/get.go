package wallet_request

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetKeysRequest struct {
	OwnerID uuid.UUID `json:"owner_id" validate:"required,uuid"`
	TOTP    string    `json:"totp" validate:"required,len=6"`
} // @name GetWalletPrivateKeysRequest

type GetSeedsRequest struct {
	OwnerID uuid.UUID `json:"owner_id" validate:"required,uuid" format:"uuid"`
	TOTP    string    `json:"totp" validate:"required,len=6"`
} // @name GetWalletSeedsRequest

type GetWalletByStoreRequest struct {
	Amount          *decimal.Decimal `json:"amount"`
	CurrencyID      *string          `json:"currency_id"`
	WalletIDs       []*uuid.UUID     `json:"wallet_ids" validate:"omitempty,dive,uuid"` //nolint:tagliatelle
	Blockchain      *string          `json:"blockchain" validate:"omitempty,oneof=bitcoin tron ethereum litecoin"`
	Address         *string          `json:"address" validate:"omitempty"`
	BalanceFiatFrom *decimal.Decimal `json:"balance_fiat_from" validate:"omitempty"`
	BalanceFiatTo   *decimal.Decimal `json:"balance_fiat_to" validate:"omitempty"`
	Page            *uint32          `json:"page" validate:"omitempty,numeric,gte=1"`
	PageSize        *uint32          `json:"page_size" validate:"omitempty,min=1,max=100"`
	StoreIDs        []uuid.UUID      `json:"store_ids" validate:"required,min=1,dive,uuid"` //nolint:tagliatelle
	IsSortByAmount  bool             `json:"is_sort_by_amount"`
	IsSortByBalance bool             `json:"is_sort_by_balance"`
} // @name GetWalletByStoreRequest

type GetHotWalletsTotalBalanceRequest struct{} // @name GetHotWalletsTotalBalanceRequest

type GetHotWalletKeysRequest struct {
	WalletAddressIDs         []uuid.UUID `json:"wallet_address_ids"`         //nolint:tagliatelle
	ExcludedWalletAddressIDs []uuid.UUID `json:"exclude_wallet_address_ids"` //nolint:tagliatelle
	TOTP                     string      `json:"totp" validate:"required,len=6"`
	FileType                 string      `json:"file_type" validate:"required,oneof=csv json txt"`
} // @name GetHotWalletKeysRequest
