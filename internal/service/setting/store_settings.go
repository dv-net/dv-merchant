package setting

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_settings"
)

const (
	ExternalWalletsListNotification = "external_wallet_email_notification"
	UserCryptoReceiptNotification   = "user_crypto_receipt_email_notification"
)

var validStoreModelSettings = map[string][]string{
	ExternalWalletsListNotification: {FlagValueDisabled, FlagValueEnabled},
	UserCryptoReceiptNotification:   {FlagValueDisabled, FlagValueEnabled},
}

type IStoreSettings interface {
	GetStoreModelSetting(ctx context.Context, name string, model IModelSetting) (*models.Setting, error)
	SetStoreModelSetting(ctx context.Context, dto UpdateDTO, opts ...repos.Option) error
	GetAvailableStoreModelSettings(ctx context.Context, model IModelSetting) ([]Dto, error)
	RemoveStoreModelSetting(ctx context.Context, store *models.Store, name string) error
}

func (s *Service) GetStoreModelSetting(ctx context.Context, name string, model IModelSetting) (*models.Setting, error) {
	if _, ok := validStoreModelSettings[name]; !ok {
		return nil, fmt.Errorf("invalid store setting name: %s", name)
	}

	arg := GetByModelParams{
		Name:      name,
		ModelID:   model.ModelID(),
		ModelType: model.ModelName(),
	}
	setting, err := s.cache.Settings().GetModelSetting(ctx, name, model)
	if err != nil {
		setting, err = s.storage.Settings().GetByModel(ctx, arg)
		if err != nil {
			return nil, err
		}
	}

	return setting, nil
}

func (s *Service) SetStoreModelSetting(ctx context.Context, dto UpdateDTO, opts ...repos.Option) error {
	if _, ok := validStoreModelSettings[dto.Name]; !ok {
		return fmt.Errorf("invalid store setting name: %s", dto.Name)
	}

	if err := s.validateStoreSetting(dto.Name, dto.Value); err != nil {
		return err
	}

	setting, err := s.storage.Settings(opts...).GetByModel(ctx, GetByModelParams{
		Name:      dto.Name,
		ModelID:   dto.Model.ModelID(),
		ModelType: dto.Model.ModelName(),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		params := repo_settings.CreateParams{
			ModelID:   dto.Model.ModelID(),
			ModelType: dto.Model.ModelName(),
			Name:      dto.Name,
			Value:     dto.Value,
			IsMutable: true,
		}
		setting, err = s.storage.Settings(opts...).Create(ctx, params)
		if err != nil {
			return fmt.Errorf("store setting initialization: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("store setting initialization: %w", err)
	}

	if setting.Value != dto.Value {
		if !setting.IsMutable {
			return errors.New("store setting is immutable")
		}

		updateParams := repo_settings.UpdateParams{
			ModelID:   dto.Model.ModelID(),
			ModelType: dto.Model.ModelName(),
			Name:      dto.Name,
			Value:     dto.Value,
			ID:        setting.ID,
			IsMutable: true,
		}
		if _, err = s.storage.Settings(opts...).Update(ctx, updateParams); err != nil {
			return fmt.Errorf("update store setting: %w", err)
		}
	}

	return nil
}

func (s *Service) GetAvailableStoreModelSettings(ctx context.Context, model IModelSetting) ([]Dto, error) {
	res, err := s.storage.Settings().GetAllByModel(ctx, repo_settings.GetAllByModelParams{
		ModelID:   model.ModelID(),
		ModelType: model.ModelName(),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch store settings: %w", err)
	}

	settings := make([]models.Setting, len(res))
	for i, setting := range res {
		settings[i] = *setting
	}

	return s.prepareStoreSettingsList(settings), nil
}

func (s *Service) RemoveStoreModelSetting(ctx context.Context, store *models.Store, name string) error {
	setting, err := s.storage.Settings().GetByModel(ctx, GetByModelParams{
		Name:      name,
		ModelID:   store.ModelID(),
		ModelType: store.ModelName(),
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("fetch store setting: %w", err)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}

	if setting == nil {
		return nil
	}

	if !setting.IsMutable {
		return errors.New("store setting is immutable")
	}

	return s.storage.Settings().DeleteByNameAndModelID(ctx, repo_settings.DeleteByNameAndModelIDParams{
		Name:    setting.Name,
		ModelID: setting.ModelID,
	})
}

func (s *Service) validateStoreSetting(name, value string) error {
	validValues, ok := validStoreModelSettings[name]
	if !ok {
		return fmt.Errorf("invalid store setting name: %s", name)
	}

	if validValues == nil {
		return nil
	}

	for _, validValue := range validValues {
		if validValue == value {
			return nil
		}
	}

	return fmt.Errorf("invalid value %s for store setting %s", value, name)
}

func (s *Service) prepareStoreSettingsList(settings []models.Setting) []Dto {
	res := make([]Dto, 0, len(settings))
	for _, setting := range settings {
		if validValues, ok := validStoreModelSettings[setting.Name]; ok {
			res = append(res, Dto{
				Name:                          setting.Name,
				Value:                         &setting.Value,
				IsEditable:                    setting.IsMutable,
				TwoFactorVerificationRequired: false,
				AvailableValues:               validValues,
			})
		}
	}
	return res
}
