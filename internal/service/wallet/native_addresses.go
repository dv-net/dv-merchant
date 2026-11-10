package wallet

import (
	"context"
	"fmt"
	"slices"

	"github.com/dv-net/dv-merchant/internal/models"
)

func (s *Service) ensureNativeAddresses(ctx context.Context, owner *models.User, wallet *models.Wallet, currencies []*models.Currency) error {
	nativeIDs := make([]string, 0)
	for _, c := range currencies {
		if c.IsFiat || c.IsNative || c.Blockchain == nil || *c.Blockchain == "" {
			continue
		}
		nativeID, err := c.Blockchain.NativeCurrency()
		if err != nil {
			continue
		}
		if slices.Contains(nativeIDs, nativeID) || slices.ContainsFunc(currencies, func(sc *models.Currency) bool {
			return sc.ID == nativeID
		}) {
			continue
		}
		nativeIDs = append(nativeIDs, nativeID)
	}
	if len(nativeIDs) == 0 {
		return nil
	}

	existing, err := s.storage.WalletAddresses().GetAllClearByWalletID(ctx, wallet.ID, nativeIDs)
	if err != nil {
		return fmt.Errorf("failed to get native wallet addresses: %w", err)
	}

	for _, nativeID := range nativeIDs {
		if slices.ContainsFunc(existing, func(wa *models.WalletAddress) bool {
			return wa.CurrencyID == nativeID
		}) {
			continue
		}

		native, err := s.storage.Currencies().GetByID(ctx, nativeID)
		if err != nil {
			return fmt.Errorf("failed to get native currency %s: %w", nativeID, err)
		}
		if !native.Status {
			continue
		}

		if _, err := s.getOrCreateWalletAddress(ctx, owner, wallet, native); err != nil {
			return fmt.Errorf("failed to create native wallet address for %s: %w", nativeID, err)
		}
	}

	return nil
}
