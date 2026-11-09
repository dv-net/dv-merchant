package aml

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_checks"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/internal/util"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ChecksWithHistoryDTO struct {
	Slug           *models.AMLSlug `json:"slug"`
	DateFrom       *string
	DateTo         *string
	Page, PageSize *uint32
}

type IHistoryService interface {
	GetCheckHistory(ctx context.Context, usr *models.User, dto ChecksWithHistoryDTO) (*storecmn.FindResponseWithFullPagination[*repo_aml_checks.FindRow], error)
	GetCheckByTransactionID(ctx context.Context, txID uuid.UUID) (*models.AmlCheck, error)
	GetBlockReasonsByTransactionID(ctx context.Context, txID uuid.UUID) (reasons models.AmlBlockReasons, blocked bool, err error)
}

var _ IHistoryService = (*Service)(nil)

func (s *Service) GetCheckHistory(ctx context.Context, usr *models.User, dto ChecksWithHistoryDTO) (*storecmn.FindResponseWithFullPagination[*repo_aml_checks.FindRow], error) {
	commonParams := storecmn.NewCommonFindParams()
	commonParams.SetPage(dto.Page)
	commonParams.SetPageSize(dto.PageSize)

	var dateFrom *time.Time
	if dto.DateFrom != nil {
		date, err := util.ParseDate(*dto.DateFrom)
		if err != nil {
			return nil, err
		}

		dateFrom = date
	}

	var dateTo *time.Time
	if dto.DateTo != nil {
		date, err := util.ParseDate(*dto.DateFrom)
		if err != nil {
			return nil, err
		}

		dateTo = date
	}

	return s.st.AmlChecks().GetByUser(ctx, usr, repo_aml_checks.GetByUserParams{
		ServiceSlug:      dto.Slug,
		DateFrom:         dateFrom,
		DateTo:           dateTo,
		CommonFindParams: *commonParams,
	})
}

func (s *Service) GetCheckByTransactionID(ctx context.Context, txID uuid.UUID) (*models.AmlCheck, error) {
	return s.st.AmlChecks().GetByTransactionID(ctx, uuid.NullUUID{UUID: txID, Valid: true})
}

// GetBlockReasonsByTransactionID returns the AML rules which blocked the transaction.
// blocked is false when the transaction is not blocked. The reasons of a blocked transaction
// are empty when it was blocked before they were recorded.
func (s *Service) GetBlockReasonsByTransactionID(ctx context.Context, txID uuid.UUID) (models.AmlBlockReasons, bool, error) {
	reasons, err := s.st.BlockedTransactions().GetBlockReasonsByTransactionID(ctx, txID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("get block reasons: %w", err)
	}

	return reasons, true, nil
}
