package transactions

import (
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/transactions_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GetUserTransactionsDTO struct {
	Currencies    []string
	StoreUuids    []uuid.UUID
	WalletAddress string
	ToAddress     string
	FromAddress   string
	Type          *models.TransactionsType
	IsSystem      bool
	MinAmountUSD  decimal.Decimal
	Blockchain    *models.Blockchain
	DateFrom      *string
	DateTo        *string
	CommonParams  *storecmn.CommonFindParams
}

func RequestToGetUserTransactionsDTO(req *transactions_request.GetByUser) GetUserTransactionsDTO {
	commonParams := storecmn.NewCommonFindParams()
	if req.Page != nil {
		commonParams.Page = req.Page
	}
	if req.PageSize != nil {
		commonParams.PageSize = req.PageSize
	}
	if req.SortBy != nil {
		commonParams.OrderBy = *req.SortBy
	} else {
		commonParams.OrderBy = "created_at_index"
	}
	if req.SortDirection != nil {
		commonParams.IsAscOrdering = *req.SortDirection == "asc"
	} else {
		commonParams.IsAscOrdering = false
	}
	return GetUserTransactionsDTO{
		Currencies:    req.Currencies,
		StoreUuids:    req.StoreUuids,
		WalletAddress: req.WalletAddress,
		ToAddress:     req.ToAddress,
		FromAddress:   req.FromAddress,
		Type:          req.Type,
		IsSystem:      req.IsSystem,
		MinAmountUSD:  req.MinAmountUSD,
		Blockchain:    req.Blockchain,
		DateFrom:      req.DateFrom,
		DateTo:        req.DateTo,
		CommonParams:  commonParams,
	}
}
