package setting

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/settings"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_settings"

	"github.com/mitchellh/mapstructure"
)

// Root settings
const (
	DvAdminSecretKey = "dv_admin_secret_key"

	ProcessingURL       = "processing_url"
	ProcessingClientID  = "processing_client_id"
	ProcessingClientKey = "processing_client_key"

	RegistrationState = "registration_state"

	MailerState           = "mailer_state"
	MailerProtocol        = "mailer_protocol"
	MailerAddress         = "mailer_address"
	MailerSender          = "mailer_sender"
	MailerUsername        = "mailer_username"
	MailerPassword        = "mailer_password"
	MailerEncryption      = "mailer_encryption"
	MerchantDomain        = "merchant_domain"
	MerchantPayFormDomain = "merchant_pay_form_domain"

	NotificationSender = "notification_sender"
	AnonymousTelemetry = "anonymous_telemetry"
)

var ExposedSettings = []string{
	ProcessingURL,
	RegistrationState,
	MailerState,
	MailerProtocol,
	MailerAddress,
	MailerSender,
	MailerUsername,
	MailerPassword,
	MailerEncryption,
	MerchantDomain,
	MerchantPayFormDomain,
	AnonymousTelemetry,
}

var MailerSettings = []string{
	MailerState,
	MailerProtocol,
	MailerAddress,
	MailerSender,
	MailerUsername,
	MailerPassword,
	MailerEncryption,
}

var SensitiveSettings = []string{
	DvAdminSecretKey,
}

var validRootSettings = map[string][]string{
	ProcessingURL:       nil,
	ProcessingClientID:  nil,
	ProcessingClientKey: nil,

	DvAdminSecretKey: nil,

	RegistrationState: {FlagValueDisabled, FlagValueEnabled},

	MailerState:           {FlagValueDisabled, FlagValueEnabled},
	MailerProtocol:        nil,
	MailerAddress:         nil,
	MailerSender:          nil,
	MailerUsername:        nil,
	MailerPassword:        nil,
	MailerEncryption:      {MailerEncryptionTypeNone, MailerEncryptionTypeTLS},
	MerchantDomain:        nil,
	MerchantPayFormDomain: nil,

	NotificationSender: {NotificationSenderInternal, NotificationSenderDVNet},
	AnonymousTelemetry: {FlagValueDisabled, FlagValueEnabled},
}

var immutableRootSettings = map[string]bool{
	ProcessingClientID:  true,
	ProcessingClientKey: true,

	ProcessingURL: true,
}

type IRootSettings interface {
	GetRootSetting(ctx context.Context, name string) (*models.Setting, error)
	SetRootSetting(ctx context.Context, name string, value string) (*models.Setting, error)
	GetRootSettings(ctx context.Context) ([]*models.Setting, error)
	GetRootSettingsByNames(ctx context.Context, names []string) ([]*models.Setting, error)
	GetRootSettingsList(ctx context.Context) ([]Dto, error)
	RemoveRootSetting(ctx context.Context, name string) error
	GetMailerSettings(ctx context.Context) (*settings.MailerSettings, error)
}

func (s *Service) GetRootSettingsList(ctx context.Context) ([]Dto, error) {
	res, err := s.storage.Settings().GetAllRootSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch root settings: %w", err)
	}

	return s.prepareSettingsList(res, true), nil
}

func (s *Service) GetRootSetting(ctx context.Context, name string) (*models.Setting, error) {
	setting, err := s.cache.Settings().GetRootSetting(ctx, name)
	if err != nil {
		setting, err = s.storage.Settings().GetByName(ctx, name)
		if err != nil {
			return nil, err
		}
		_, _ = s.cache.Settings().SetRootSetting(ctx, setting)
	}
	return setting, nil
}

func (s *Service) GetRootSettings(ctx context.Context) ([]*models.Setting, error) {
	settingList, err := s.cache.Settings().GetRootSettings(ctx)
	if err != nil || len(settingList) != len(validRootSettings) {
		settingList, err = s.storage.Settings().GetAllRootSettings(ctx)
		if err != nil {
			return nil, err
		}
		for _, setting := range settingList {
			_, _ = s.cache.Settings().SetRootSetting(ctx, setting)
		}
	}
	return settingList, nil
}

func (s *Service) RemoveRootSetting(ctx context.Context, name string) error {
	setting, err := s.cache.Settings().GetRootSetting(ctx, name)
	if err != nil {
		setting, err = s.storage.Settings().GetByName(ctx, name)
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

func (s *Service) SetRootSetting(ctx context.Context, name, value string) (*models.Setting, error) {
	_, ok := validRootSettings[name]
	if !ok {
		return nil, fmt.Errorf("invalid setting name: %s", name)
	}
	if err := s.validateSetting(true, name, value); err != nil {
		return nil, fmt.Errorf("invalid setting value: %w", err)
	}

	setting, err := s.cache.Settings().GetRootSetting(ctx, name)
	isImmutable := immutableRootSettings[name]
	if err != nil || setting == nil {
		setting, err = s.storage.Settings().GetByName(ctx, name)
		if err != nil {
			params := repo_settings.CreateParams{Name: name, Value: value, IsMutable: !isImmutable}
			setting, err = s.storage.Settings().Create(ctx, params)
			if err != nil {
				return nil, err
			}
		}

		_, _ = s.cache.Settings().SetRootSetting(ctx, setting)
	}

	if setting.Value != value {
		if !setting.IsMutable {
			return nil, errors.New("setting is immutable")
		}

		updateParams := repo_settings.UpdateParams{
			ID:        setting.ID,
			Name:      name,
			Value:     value,
			IsMutable: !isImmutable,
		}
		setting, err = s.storage.Settings().Update(ctx, updateParams)
		if err != nil {
			return nil, err
		}
		_, _ = s.cache.Settings().SetRootSetting(ctx, setting)
	}

	if slices.Index(ExposedSettings, name) != -1 {
		s.notifySettingChange(name, value)
	}

	if slices.Index(MailerSettings, name) != -1 {
		s.notifyMailerSettingsChange(name, value)
	}

	if name == NotificationSender {
		s.notifyNotificationSenderChanged(value)
	}

	return setting, nil
}

func (s *Service) GetMailerSettings(ctx context.Context) (*settings.MailerSettings, error) {
	sets, err := s.GetRootSettingsByNames(ctx, MailerSettings)
	if err != nil {
		return nil, err
	}

	input := make(map[string]interface{}, len(MailerSettings))
	for _, set := range sets {
		input[set.Name] = set.Value
	}

	mailerSettings := settings.MailerSettings{}
	if err := mapstructure.Decode(input, &mailerSettings); err != nil {
		return nil, err
	}

	return &mailerSettings, nil
}

func (s *Service) GetRootSettingsByNames(ctx context.Context, names []string) ([]*models.Setting, error) {
	return s.storage.Settings().GetByNames(
		ctx,
		names,
	)
}
