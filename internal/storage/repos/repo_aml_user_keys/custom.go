package repo_aml_user_keys

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/dv-net/dv-processing/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ICustomQuerier interface {
	Querier
	GetServiceCredentials(ctx context.Context, userID uuid.UUID, slug models.AMLSlug) (*ServiceCredsResult, error)
}

type CustomQueries struct {
	*Queries
	psql DBTX
}

func NewCustom(psql DBTX) *CustomQueries {
	return &CustomQueries{
		Queries: New(psql),
		psql:    psql,
	}
}

func (s *CustomQueries) WithTx(tx pgx.Tx) *CustomQueries {
	return &CustomQueries{
		Queries: New(tx),
		psql:    tx,
	}
}

// ServiceCredentials represents a map of credential names to their values.
type ServiceCredentials map[models.AmlKeyType]string

type ServiceCredsResult struct {
	Service *models.AmlService
	Creds   ServiceCredentials
}

// GetServiceCredentials fetches credentials for a user and service slug.
func (s *CustomQueries) GetServiceCredentials(ctx context.Context, userID uuid.UUID, slug models.AMLSlug) (*ServiceCredsResult, error) {
	rows, err := s.PrepareServiceDataByUserAndSlug(ctx, slug, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credentials: %w", err)
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("no credentials found for user %s and slug %s", userID, slug)
	}

	creds := make(ServiceCredentials, len(rows))
	result := &ServiceCredsResult{
		Service: utils.Pointer(rows[0].AmlService),
		Creds:   creds,
	}

	for _, row := range rows {
		// Check if all keys belong to one service
		if row.AmlService.ID != result.Service.ID {
			return nil, fmt.Errorf("invalid service data: expected ID %d and slug %s, got ID %d and slug %s",
				result.Service.ID, result.Service.Slug, row.AmlService.ID, row.AmlService.Slug)
		}

		creds[row.Name] = row.Value
	}

	return result, nil
}
