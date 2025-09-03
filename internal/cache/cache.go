package cache

import (
	"github.com/dv-net/dv-merchant/internal/cache/settings"
	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/jellydator/ttlcache/v3"
)

type ICache interface {
	Settings() settings.ISettingsCache
}

type cache struct {
	settings settings.ISettingsCache
}

func (o cache) Settings() settings.ISettingsCache {
	return o.settings
}

func InitCache() ICache {
	return &cache{
		settings: settings.New(ttlcache.New[string, *models.Setting]()),
	}
}
