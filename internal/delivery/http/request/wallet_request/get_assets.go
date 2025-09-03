package wallet_request

import (
	"github.com/shopspring/decimal"

	"github.com/dv-net/dv-merchant/internal/models"
)

type GetWalletAssetsRequest struct {
	Blockchains []models.Blockchain `json:"blockchains" validate:"unique"`
	Currencies  []string            `json:"currencies" validate:"unique"`
} // @name GetWalletAssetsRequest

type GetSummarizedUserWalletsRequest struct {
	MinBalance decimal.Decimal `json:"min_balance" query:"min_balance"`
} // @name GetSummarizedUserWalletsRequest
