package store

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_store_aml_settings"
	"github.com/dv-net/dv-merchant/pkg/dbutils/pgerror"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type IStoreAmlSettings interface {
	GetStoreAmlSettings(ctx context.Context, storeID uuid.UUID) (*models.StoreAmlSetting, error)
	UpdateAMLSetting(ctx context.Context, storeID uuid.UUID, dto UpdateAMLSettingsDTO) (*models.StoreAmlSetting, error)
}

func (s *Service) GetStoreAmlSettings(ctx context.Context, storeID uuid.UUID) (*models.StoreAmlSetting, error) {
	amlSetting, err := s.storage.StoreAmlSettings().GetByStoreID(ctx, storeID)
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.log.Debug("error get aml settings", parsedErr)
		return nil, parsedErr
	}
	return amlSetting, nil
}

func (s *Service) UpdateAMLSetting(ctx context.Context, storeID uuid.UUID, dto UpdateAMLSettingsDTO) (*models.StoreAmlSetting, error) {
	amlSetting, err := s.storage.StoreAmlSettings().Upsert(ctx, repo_store_aml_settings.UpsertParams{
		StoreID:       storeID,
		Enabled:       dto.Enabled,
		RiskThreshold: dto.RiskThreshold,
		ProviderSlug:  dto.ProviderSlug,
	})
	if err != nil {
		parsedErr := pgerror.ParseError(err)
		s.log.Debug("error mark address is dirty", parsedErr)
		return nil, parsedErr
	}
	return amlSetting, nil
}

func isScoreAboveThreshold(score decimal.Decimal, threshold int32) bool {
	return score.GreaterThanOrEqual(decimal.NewFromInt32(threshold))
}
