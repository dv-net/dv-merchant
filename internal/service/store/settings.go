package store

import (
	"context"

	"github.com/google/uuid"

	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
)

func (s *Service) CreateStoreSettings(ctx context.Context, storeID uuid.UUID, opts ...repos.Option) error {
	store, err := s.GetStoreByID(ctx, storeID)
	if err != nil {
		return err
	}

	settings := []setting.UpdateDTO{
		{
			Name:  setting.ExternalWalletsListNotification,
			Value: setting.FlagValueEnabled,
			Model: store,
		},
		{
			Name:  setting.UserCryptoReceiptNotification,
			Value: setting.FlagValueEnabled,
			Model: store,
		},
	}

	for _, sDTO := range settings {
		if err := s.settingSvc.SetStoreModelSetting(ctx, sDTO, opts...); err != nil {
			return err
		}
	}

	return nil
}
