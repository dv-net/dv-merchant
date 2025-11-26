package transactions_request

import (
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetByUser struct {
	Currencies    []string                 `json:"currencies" query:"currencies"`
	StoreUuids    []uuid.UUID              `json:"store_uuids" query:"store_uuids"`
	WalletAddress string                   `json:"wallet_address" query:"wallet_address"`
	ToAddress     string                   `json:"to_address" query:"to_address"`
	FromAddress   string                   `json:"from_address" query:"from_address"`
	Type          *models.TransactionsType `json:"type" query:"type" validate:"omitempty,oneof=deposit transfer" enums:"deposit,transfer"`
	IsSystem      bool                     `json:"is_system" query:"is_system"`
	MinAmountUSD  decimal.Decimal          `json:"min_amount_usd" query:"min_amount_usd"`
	Blockchain    *models.Blockchain       `json:"blockchain" query:"blockchain"`
	DateFrom      *string                  `json:"date_from" query:"date_from" format:"date-time"`
	DateTo        *string                  `json:"date_to" query:"date_to" format:"date-time"`
	Page          *uint32                  `json:"page" query:"page"`
	PageSize      *uint32                  `json:"page_size" query:"page_size"`
	SortBy        *string                  `json:"sort_by" query:"sort_by,default:created_at_index" enums:"created_at_index,amount_usd,tx_hash,user_email" oneof:"created_at_index,amount_usd,tx_hash,user_email"`
	SortDirection *string                  `json:"sort_direction" query:"sort_direction,default:desc" enums:"asc,desc" oneof:"asc,desc"`
} //	@name	GetTransactionsByUserRequest

type GetByUserExported struct {
	Format        string                   `json:"format" query:"format" validate:"required,oneof=csv xlsx"`
	Currencies    []string                 `json:"currencies" query:"currencies"`
	StoreUuids    []uuid.UUID              `json:"store_uuids" query:"store_uuids"`
	WalletAddress string                   `json:"wallet_address" query:"wallet_address"`
	ToAddress     string                   `json:"to_address" query:"to_address"`
	FromAddress   string                   `json:"from_address" query:"from_address"`
	Type          *models.TransactionsType `json:"type" query:"type" validate:"omitempty,oneof=deposit transfer" enums:"deposit,transfer"`
	IsSystem      bool                     `json:"is_system" query:"is_system"`
	MinAmountUSD  decimal.Decimal          `json:"min_amount_usd" query:"min_amount_usd"`
	Blockchain    *models.Blockchain       `json:"blockchain" query:"blockchain"`
	DateFrom      *string                  `json:"date_from" query:"date_from" format:"date"`
	DateTo        *string                  `json:"date_to" query:"date_to" format:"date"`
	Page          *uint32                  `json:"page" query:"page"`
	PageSize      *uint32                  `json:"page_size" query:"page_size"`
} //	@name	GetTransactionsByUserExportedRequest
