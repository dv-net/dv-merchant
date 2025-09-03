package setting

import (
	"fmt"
	"slices"
	"sort"

	"github.com/dv-net/dv-merchant/internal/cache"
	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/event"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_settings"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/google/uuid"
)

type IModelSetting interface {
	ModelID() uuid.NullUUID
	ModelName() *string
}

type ISettingService interface {
	IUserSettings
	IRootSettings

	Is2faRequired(name string) bool
}

type Service struct {
	cfg           *config.Config
	storage       storage.IStorage
	cache         cache.ICache
	eventListener event.IListener
	log           logger.Logger
}

type GetByModelParams = repo_settings.GetByModelParams

func New(
	cfg *config.Config,
	storage storage.IStorage,
	cache cache.ICache,
	eventListener event.IListener,
	log logger.Logger,
) ISettingService {
	return &Service{
		cfg:           cfg,
		storage:       storage,
		cache:         cache,
		eventListener: eventListener,
		log:           log,
	}
}

func (s *Service) prepareSettingsList(existingSettings []*models.Setting, isRoot bool) []Dto {
	currentValues := make(map[string]*models.Setting, len(existingSettings))
	for _, val := range existingSettings {
		currentValues[val.Name] = val
	}

	validSettings := validModelSettings
	if isRoot {
		validSettings = validRootSettings
	}
	preparedRes := make([]Dto, 0, len(validSettings))
	for name, availValues := range validSettings {
		if s.isSensitiveSetting(name) {
			continue
		}

		var currentValue *string
		isEditable := true // user setting is always editable
		if isRoot {
			if isRootEditable, ok := immutableRootSettings[name]; ok {
				isEditable = isRootEditable
			}
		}

		if value, ok := currentValues[name]; ok {
			currentValue = &value.Value
			isEditable = value.IsMutable
		}

		preparedRes = append(preparedRes, Dto{
			Name:                          name,
			Value:                         currentValue,
			IsEditable:                    isEditable,
			TwoFactorVerificationRequired: s.Is2faRequired(name),
			AvailableValues:               availValues,
		})
	}

	sort.Sort(ByName(preparedRes))
	return preparedRes
}

func (s *Service) validateSetting(isRoot bool, name, value string) error {
	targetSet := validModelSettings
	if isRoot {
		targetSet = validRootSettings
	}

	availableValues, ok := targetSet[name]
	if !ok {
		return fmt.Errorf("settings with name '%s' is not available", name)
	}

	valueIsAvailable := true
	if len(availableValues) > 0 {
		valueIsAvailable = false
		for _, val := range availableValues {
			if val == value {
				valueIsAvailable = true
				break
			}
		}
	}

	if !valueIsAvailable {
		return fmt.Errorf("unexpected value '%s' for setting '%s'", value, name)
	}

	return nil
}

func (s *Service) Is2faRequired(name string) bool {
	ffaSettings := map[string]struct{}{
		WithdrawFromProcessing: {},
	}

	_, ok := ffaSettings[name]
	return ok
}

func (s *Service) isSensitiveSetting(name string) bool {
	return slices.Contains(SensitiveSettings, name)
}
