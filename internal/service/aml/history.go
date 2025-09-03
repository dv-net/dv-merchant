package aml

import (
	"context"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_aml_checks"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/internal/util"
)

type ChecksWithHistoryDTO struct {
	Slug           *models.AMLSlug `json:"slug"`
	DateFrom       *string
	DateTo         *string
	Page, PageSize *uint32
}

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
