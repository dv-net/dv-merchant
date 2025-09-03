package user

import (
	"context"
	"errors"
	"slices"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/setting"
)

type SettingUpdateDTO struct {
	OTP   string
	Name  string
	Value *string
	Model setting.IModelSetting
}

type InitSettingUpdateDTO struct {
	Name  string
	Model setting.IModelSetting
}

type IUserSettings interface {
	SettingUpdate(ctx context.Context, usr *models.User, dto SettingUpdateDTO) error
}

var _ IUserSettings = (*Service)(nil)

func (s *Service) SettingUpdate(ctx context.Context, usr *models.User, dto SettingUpdateDTO) error {
	if s.settingsService.Is2faRequired(dto.Name) {
		if err := s.ensure2faEnabledByUser(ctx, usr); err != nil {
			return err
		}

		if err := s.processingService.ValidateTwoFactorToken(ctx, usr.ProcessingOwnerID.UUID, dto.OTP); err != nil {
			return ErrInvalidOTP
		}
	}

	if dto.Model != nil && dto.Model.ModelName() != nil {
		return s.processUserSettingUpdate(ctx, usr, dto.Name, dto.Value, dto.Model)
	}

	return s.processRootSettingUpdate(ctx, usr, dto.Name, dto.Value)
}

func (s *Service) processRootSettingUpdate(
	ctx context.Context,
	usr *models.User,
	settingName string,
	newValue *string,
) error {
	roles, err := s.permissionService.UserRoles(usr.ID.String())
	if err != nil {
		return err
	}

	if !slices.Contains(roles, models.UserRoleRoot) {
		return errors.New("only root user can update setting")
	}

	if newValue == nil {
		return s.settingsService.RemoveRootSetting(ctx, settingName)
	}

	_, err = s.settingsService.SetRootSetting(ctx, settingName, *newValue)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) processUserSettingUpdate(
	ctx context.Context,
	usr *models.User,
	settingName string,
	newValue *string,
	settingTarget setting.IModelSetting,
) error {
	if !settingTarget.ModelID().Valid {
		return errors.New("invalid user data")
	}

	if newValue == nil {
		return s.settingsService.RemoveSetting(ctx, usr, settingName)
	}

	return s.settingsService.SetModelSetting(ctx, setting.UpdateDTO{
		Name:  settingName,
		Value: *newValue,
		Model: settingTarget,
	})
}

func (s *Service) ensure2faEnabledByUser(ctx context.Context, usr *models.User) error {
	if !usr.ProcessingOwnerID.Valid {
		return ErrOwnerIDIsNotSet
	}

	twoFactorData, err := s.processingService.GetTwoFactorAuthData(ctx, usr.ProcessingOwnerID.UUID)
	if err != nil {
		return err
	}

	if !twoFactorData.IsConfirmed {
		return ErrTwoFactorAuthIsNotConfirmed
	}

	return nil
}
