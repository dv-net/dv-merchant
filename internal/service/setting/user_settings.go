package setting

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/dv-net/dv-merchant/internal/cache/settings"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_settings"
)

// TransfersStatus Common settings
const (
	TransfersStatus = "transfers_status"
)

// TransferType Tron settings
const (
	TransferType = "transfer_type"
)

const (
	QuickStartGuideStatus = "quick_start_guide_status"
)

const (
	WithdrawFromProcessing = "withdraw_from_processing"
)

var validModelSettings = map[string][]string{
	TransfersStatus:        {FlagValueDisabled, FlagValueEnabled, TransferStatusSystemSuspended},
	TransferType:           {string(TransferByBurnTRX), string(TransferByResource), string(TransferByCloudDelegate)},
	QuickStartGuideStatus:  {FlagValueIncompleted, FlagValueCompleted},
	WithdrawFromProcessing: {FlagValueDisabled, FlagValueEnabled},
}

type IUserSettings interface {
	GetModelSetting(ctx context.Context, name string, model IModelSetting) (*models.Setting, error)
	SetModelSetting(ctx context.Context, dto UpdateDTO, option ...repos.Option) error
	GetAvailableModelSettings(ctx context.Context, model IModelSetting) ([]Dto, error)
	RemoveSetting(ctx context.Context, user *models.User, name string) error
}

func (s *Service) GetAvailableModelSettings(ctx context.Context, model IModelSetting) ([]Dto, error) {
	res, err := s.storage.Settings().GetAllByModel(ctx, repo_settings.GetAllByModelParams{
		ModelID:   model.ModelID(),
		ModelType: model.ModelName(),
	})
	if err != nil {
		return nil, fmt.Errorf("fetch settings: %w", err)
	}

	return s.prepareSettingsList(res, false), nil
}

func (s *Service) notifySettingChange(name, value string) {
	ev := &RootSettingChangedEvent{
		SettingName:  name,
		SettingValue: value,
	}
	err := s.eventListener.Fire(ev)
	if err != nil {
		s.log.Error("fire event", err)
	}
}

func (s *Service) notifyMailerSettingsChange(name, value string) {
	ev := &MailerSettingChangedEvent{
		SettingName:  name,
		SettingValue: value,
	}
	err := s.eventListener.Fire(ev)
	if err != nil {
		s.log.Error("fire event", err)
	}
}

func (s *Service) notifyNotificationSenderChanged(newValue string) {
	ev := NotificationSenderChangedEvent{newValue}
	err := s.eventListener.Fire(ev)
	if err != nil {
		s.log.Error("fire event", err)
	}
}

func (s *Service) GetModelSetting(ctx context.Context, name string, model IModelSetting) (*models.Setting, error) {
	if _, ok := validModelSettings[name]; !ok {
		return nil, fmt.Errorf("invalid setting name: %s", name)
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
		_, _ = s.cache.Settings().SetRootSetting(ctx, setting)
	}

	return setting, nil
}

type UpdateDTO struct {
	Name  string
	Value string
	Model IModelSetting
}

func (s *Service) SetModelSetting(ctx context.Context, dto UpdateDTO, opts ...repos.Option) error {
	if _, ok := validModelSettings[dto.Name]; !ok {
		return fmt.Errorf("invalid setting name: %s", dto.Name)
	}

	if err := s.validateSetting(false, dto.Name, dto.Value); err != nil {
		return err
	}

	setting, err := s.cache.Settings().GetModelSetting(ctx, dto.Name, dto.Model)
	if err != nil || setting == nil {
		setting, err = s.storage.Settings(opts...).GetByModel(ctx, GetByModelParams{
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
				return fmt.Errorf("setting initialization: %w", err)
			}
		}
		if err != nil {
			return fmt.Errorf("setting initialization: %w", err)
		}

		_, _ = s.cache.Settings().SetRootSetting(ctx, setting)
	}

	if setting.Value != dto.Value {
		if !setting.IsMutable {
			return errors.New("setting is immutable")
		}

		updateParams := repo_settings.UpdateParams{
			ModelID:   dto.Model.ModelID(),
			ModelType: dto.Model.ModelName(),
			Name:      dto.Name,
			Value:     dto.Value,
			ID:        setting.ID,
			IsMutable: true,
		}
		setting, err = s.storage.Settings(opts...).Update(ctx, updateParams)
		if err != nil {
			return fmt.Errorf("update setting: %w", err)
		}
		_, _ = s.cache.Settings().SetRootSetting(ctx, setting)
	}

	return nil
}

type RemoveDTO struct {
	Name string
	Code int
}

func (s *Service) RemoveSetting(ctx context.Context, user *models.User, name string) error {
	setting, err := s.cache.Settings().GetModelSetting(ctx, name, settings.IModelSetting(user))
	if err != nil {
		setting, err = s.storage.Settings().GetByModel(ctx, GetByModelParams{
			Name:      name,
			ModelID:   user.ModelID(),
			ModelType: user.ModelName(),
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("fetch setting: %w", err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			// setting is already removed
			return nil
		}
	}

	if setting == nil {
		return nil
	}

	if !setting.IsMutable {
		return errors.New("setting is immutable")
	}

	s.cache.Settings().RemoveRootSetting(ctx, setting.Name)
	return s.storage.Settings().DeleteRootByName(ctx, setting.Name)
}
