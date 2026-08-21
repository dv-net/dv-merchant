package aml

import (
	"context"
	"fmt"
	"time"

	amldto "github.com/dv-net/dv-merchant/internal/service/aml/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) GetStatistics(ctx context.Context, userID uuid.UUID) (*amldto.Statistics, error) {
	now := time.Now()
	dateFrom := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dateTo := dateFrom.AddDate(0, 0, 1)

	statistics, err := s.st.AmlChecks().GetStatistics(
		ctx,
		userID,
		pgtype.Timestamp{Time: dateFrom.UTC(), Valid: true},
		pgtype.Timestamp{Time: dateTo.UTC(), Valid: true},
	)
	if err != nil {
		return nil, fmt.Errorf("get AML statistics: %w", err)
	}

	return &amldto.Statistics{
		CheckedToday:    statistics.CheckedToday,
		SuccessfulToday: statistics.SuccessfulToday,
		FailedToday:     statistics.FailedToday,
	}, nil
}
