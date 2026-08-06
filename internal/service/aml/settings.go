package aml

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_user_aml_settings"
	"github.com/dv-net/dv-merchant/pkg/dbutils/pgerror"
	"github.com/google/uuid"
)

type IUserAmlSettings interface {
	GetAmlSettings(ctx context.Context, userID uuid.UUID) (*models.UserAmlSetting, error)
	UpdateAmlSettings(ctx context.Context, userID uuid.UUID, dto UpdateAmlSettingsDTO) (*models.UserAmlSetting, error)
	ListRiskRules(ctx context.Context, userID uuid.UUID, providerSlug *models.AMLSlug) ([]*models.UserAmlRiskRule, error)
	UpsertRiskRules(ctx context.Context, userID uuid.UUID, providerSlug *models.AMLSlug, rules []RiskRuleDTO) ([]*models.UserAmlRiskRule, error)
}

var _ IUserAmlSettings = (*Service)(nil)

func (s *Service) GetAmlSettings(ctx context.Context, userID uuid.UUID) (*models.UserAmlSetting, error) {
	settings, err := s.st.UserAmlSettings().GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *Service) UpdateAmlSettings(ctx context.Context, userID uuid.UUID, dto UpdateAmlSettingsDTO) (*models.UserAmlSetting, error) {
	settings, err := s.st.UserAmlSettings().UpsertAmlSetting(ctx, repo_user_aml_settings.UpsertAmlSettingParams{
		UserID:       userID,
		Enabled:      dto.Enabled,
		ProviderSlug: dto.ProviderSlug,
	})

	if err != nil {
		s.log.Errorw("could not update settings", "error", err, "user_id", userID, "enabled", dto.Enabled, "provider_slug", dto.ProviderSlug)
		return nil, pgerror.ParseError(err)
	}

	return settings, nil
}

func (s *Service) ListRiskRules(ctx context.Context, userID uuid.UUID, providerSlug *models.AMLSlug) ([]*models.UserAmlRiskRule, error) {
	rules, err := s.st.UserAmlSettings().ListRiskRulesByUserID(ctx, repo_user_aml_settings.ListRiskRulesByUserIDParams{
		UserID:       userID,
		ProviderSlug: providerSlug.String(),
	})

	if err != nil {
		return nil, pgerror.ParseError(err)
	}

	return rules, nil
}

func (s *Service) UpsertRiskRules(ctx context.Context, userID uuid.UUID, providerSlug *models.AMLSlug, rules []RiskRuleDTO) ([]*models.UserAmlRiskRule, error) {
	result := make([]*models.UserAmlRiskRule, 0, len(rules))
	for _, r := range rules {
		rule, err := s.st.UserAmlSettings().UpsertAmlRiskRule(ctx, repo_user_aml_settings.UpsertAmlRiskRuleParams{
			UserID:       userID,
			ProviderSlug: providerSlug.String(),
			RiskType:     r.RiskType,
			Enabled:      r.Enabled,
			Threshold:    r.Threshold,
			Action:       r.Action,
		})

		if err != nil {
			return nil, pgerror.ParseError(err)
		}
		result = append(result, rule)
	}
	return result, nil
}
