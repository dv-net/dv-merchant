package public_request

import (
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
)

type TopUpFormByStoreRequest struct {
	IP     *string `json:"ip,omitempty" query:"ip" validate:"omitempty,ip" format:"ipv4"`
	Email  *string `json:"email,omitempty" query:"email" validate:"omitempty" format:"email"`
	Locale *string `json:"locale,omitempty" query:"locale" validate:"omitempty"`
} // @name TopUpFormByStoreRequest

type StoreDto struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	CurrencyID     string    `json:"currency_id"`
	SiteURL        *string   `json:"site_url,omitempty"`
	ReturnURL      *string   `json:"return_url,omitempty"`
	SuccessURL     *string   `json:"success_url,omitempty"`
	MinimalPayment string    `json:"minimal_payment"`
	Status         bool      `json:"status"`
} // @name PublicStore

type CurrencyDto struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Code            string            `json:"code"`
	Blockchain      models.Blockchain `json:"blockchain"`
	CurrencyLabel   *string           `json:"currency_label"`
	TokenLabel      *string           `json:"token_label"`
	IsNative        bool              `json:"is_native"`
	ContractAddress *string           `json:"contract_address"`
	Order           int64             `json:"order"`
} // @name PublicCurrency

type WalletAddressDto struct {
	Currency CurrencyDto `json:"currency"`
	Address  string      `json:"address"`
} // @name PublicWalletAddress

type GetWalletRequest struct {
	Locale     *string `json:"locale,omitempty" query:"locale" validate:"omitempty"`
	CurrencyID *string `json:"currency_id,omitempty" query:"currency_id" validate:"omitempty"`
} // @name GetWalletRequest

type GetWalletDto struct {
	Store     StoreDto           `json:"store"`
	WalletID  uuid.UUID          `json:"wallet_id"`
	Addresses []WalletAddressDto `json:"addresses"`
	Rates     map[string]string  `json:"rates"`
} // @name PublicGetWalletResponse
