package currency_response

import (
	"github.com/shopspring/decimal"
)

type GetCurrencyResponse struct {
	ID                   string          `json:"id"`
	Code                 string          `json:"code"`
	Name                 string          `json:"name"`
	Precision            int             `json:"precision"`
	IsFiat               bool            `json:"is_fiat"`
	Blockchain           string          `json:"blockchain"`
	ContractAddress      string          `json:"contract_address"`
	WithdrawalMinBalance decimal.Decimal `json:"withdrawal_min_balance"`
	HasBalance           bool            `json:"has_balance"`
	Status               bool            `json:"status"`
	MinConfirmation      int             `json:"min_confirmation"`
	Icon                 CurrencyIcon    `json:"icon"`
	BlockchainIcon       CurrencyIcon    `json:"blockchain_icon"`
	ExplorerLink         string          `json:"explorer_link"`
} // @name GetCurrencyResponse

type CurrencyIcon struct {
	Icon128 string `json:"icon_128"` //nolint:tagliatelle
	Icon512 string `json:"icon_512"` //nolint:tagliatelle
	IconSVG string `json:"icon_svg"`
} // @name CurrencyIcon
