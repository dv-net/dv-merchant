package settings

import (
	"context"

	"github.com/dv-net/dv-merchant/internal/cache/errors"
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
)

type IModelSetting interface {
	ModelID() uuid.NullUUID
	ModelName() *string
}

type ISettingsCache interface {
	GetRootSetting(ctx context.Context, name string) (*models.Setting, error)
	SetRootSetting(ctx context.Context, value *models.Setting) (*models.Setting, error)
	GetRootSettings(ctx context.Context) ([]*models.Setting, error)
	GetModelSetting(ctx context.Context, name string, model IModelSetting) (*models.Setting, error)
	RemoveRootSetting(_ context.Context, name string)
}

type settings struct {
	client *ttlcache.Cache[string, *models.Setting]
}

func (o settings) GetRootSetting(_ context.Context, name string) (*models.Setting, error) {
	entry := o.client.Get(name)
	if entry == nil {
		return nil, errors.ErrEntryNotFound
	}
	return entry.Value(), nil
}

func (o settings) SetRootSetting(_ context.Context, value *models.Setting) (*models.Setting, error) {
	entry := o.client.Set(value.Name, value, ttlcache.NoTTL)
	return entry.Value(), nil
}

func (o settings) GetRootSettings(_ context.Context) ([]*models.Setting, error) {
	entries := o.client.Items()
	settings := make([]*models.Setting, 0, len(entries))
	for _, entry := range entries {
		settings = append(settings, entry.Value())
	}
	if len(settings) == 0 {
		return nil, errors.ErrEntryNotFound
	}
	return settings, nil
}

func (o settings) GetModelSetting(_ context.Context, name string, model IModelSetting) (*models.Setting, error) {
	entries := o.client.Items()
	for _, entry := range entries {
		if entry.Value().ModelID == model.ModelID() && entry.Value().ModelType == model.ModelName() && entry.Value().Name == name {
			return entry.Value(), nil
		}
	}
	return nil, errors.ErrEntryNotFound
}

func (o settings) RemoveRootSetting(_ context.Context, name string) {
	o.client.Delete(name)
}

func New(client *ttlcache.Cache[string, *models.Setting]) ISettingsCache {
	return &settings{
		client: client,
	}
}

var _ ISettingsCache = (*settings)(nil)
