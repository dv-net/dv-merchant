package refund

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_refund_requests"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/internal/util"

	"github.com/google/uuid"
)

type FindDTO struct {
	UserID             uuid.UUID
	Statuses           []string
	Blockchain         *models.Blockchain
	TxHash             string
	FromAddress        string
	DestinationAddress string
	DateFrom           *string
	DateTo             *string
	Page               *uint32
	PageSize           *uint32
	SortBy             *string
	SortDirection      *string
}

// FindByUser lists refund requests of all stores owned by the user with filters and pagination.
// Each row carries the blocked deposit, the AML check that blocked it and the refund transfer.
func (s *Service) FindByUser(ctx context.Context, dto FindDTO) (*storecmn.FindResponseWithFullPagination[*repo_refund_requests.FindRow], error) {
	params := repo_refund_requests.FindParams{
		UserID:             dto.UserID,
		Statuses:           dto.Statuses,
		Blockchain:         dto.Blockchain,
		TxHash:             dto.TxHash,
		FromAddress:        dto.FromAddress,
		DestinationAddress: dto.DestinationAddress,
	}
	params.Page = dto.Page
	params.PageSize = dto.PageSize

	if dto.SortBy != nil {
		params.OrderBy = *dto.SortBy
	}
	params.IsAscOrdering = dto.SortDirection != nil && *dto.SortDirection == "asc"

	if dto.DateFrom != nil && *dto.DateFrom != "" {
		dateFrom, err := util.ParseDate(*dto.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from: %w", err)
		}
		params.DateFrom = dateFrom
	}

	if dto.DateTo != nil && *dto.DateTo != "" {
		dateTo, err := util.ParseDate(*dto.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to: %w", err)
		}
		params.DateTo = dateTo
	}

	return s.storage.RefundRequests().Find(ctx, params)
}
