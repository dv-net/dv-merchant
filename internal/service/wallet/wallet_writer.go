package wallet

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallets"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"
	"github.com/jackc/pgx/v5"
)

// StoreWalletWithAddress creates/returns wallet with addresses
func (s *Service) StoreWalletWithAddress(ctx context.Context, dto CreateStoreWalletWithAddressDTO, amountUSD string) (*WithAddressDto, error) {
	var storeOwner *models.User
	var walletEmail *string
	walletWithAddress := &WithAddressDto{}
	wallet := &models.Wallet{}

	err := repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		w, err := s.storage.Wallets(repos.WithTx(tx)).GetByStore(ctx, repo_wallets.GetByStoreParams{
			StoreID:         dto.StoreID,
			StoreExternalID: dto.StoreExternalID,
		})
		if err != nil {
			w, err = s.storage.Wallets(repos.WithTx(tx)).Create(ctx, dto.ToCreateParams())
			if err != nil {
				return err
			}
		}

		if w.Email.Valid {
			walletEmail = &w.Email.String
		}
		if err = s.updateWalletMeta(ctx, w, dto.ToCreateParams(), &walletEmail, repos.WithTx(tx)); err != nil {
			return err
		}

		wallet = w
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = repos.BeginTxFunc(ctx, s.storage.PSQLConn(), pgx.TxOptions{}, func(tx pgx.Tx) error {
		feURL, err := s.settingService.GetRootSetting(ctx, setting.MerchantPayFormDomain)
		if err != nil {
			return err
		}

		if err := walletWithAddress.Encode(wallet, feURL.Value); err != nil {
			return fmt.Errorf("failed to encode wallet: %w", err)
		}

		str, err := s.storage.Stores().GetByID(ctx, dto.StoreID)
		if err != nil {
			return err
		}

		storeOwner, err = s.storage.Users().GetByID(ctx, str.UserID)
		if err != nil {
			return err
		}

		if ownerID := storeOwner.ProcessingOwnerID; !ownerID.Valid {
			return errors.New("store owner processing uuid is not valid")
		}

		currencies, err := s.storage.StoreCurrencies().GetAllByStoreID(ctx, str.ID)
		if err != nil {
			return err
		}

		address, err := s.generateWalletAddresses(ctx, tx, storeOwner, wallet, str, currencies, amountUSD)
		if err != nil {
			return err
		}
		walletWithAddress.Address = address

		rates, err := s.exrateService.GetStoreCurrencyRate(ctx, currencies, str.RateSource.String(), str.RateScale)
		if err != nil {
			return fmt.Errorf("failed to get store currency rate: %w", err)
		}
		walletWithAddress.Rates = rates
		walletWithAddress.AmountUSD = amountUSD

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to store wallet with address for store external id %s: %w", dto.StoreExternalID, err)
	}

	return walletWithAddress, nil
}

func (s *Service) updateWalletMeta(ctx context.Context, wallet *models.Wallet, params repo_wallets.CreateParams, emailPtr **string, tx ...repos.Option) error {
	if params.UntrustedEmail.Valid && (!wallet.UntrustedEmail.Valid || params.UntrustedEmail.String != wallet.UntrustedEmail.String) {
		if err := s.storage.Wallets(tx...).UpdateUserUntrustedEmail(ctx, repo_wallets.UpdateUserUntrustedEmailParams{
			UntrustedEmail: params.UntrustedEmail,
			ID:             wallet.ID,
		}); err != nil {
			return err
		}
		*emailPtr = &params.UntrustedEmail.String
	}

	if params.Email.Valid && (!wallet.Email.Valid || params.Email.String != wallet.Email.String) {
		if err := s.storage.Wallets(tx...).UpdateUserEmail(ctx, repo_wallets.UpdateUserEmailParams{
			Email: params.Email,
			ID:    wallet.ID,
		}); err != nil {
			return err
		}
		*emailPtr = &params.Email.String
	}

	if params.IpAddress.Valid && (!wallet.IpAddress.Valid || params.IpAddress.String != wallet.IpAddress.String) {
		if err := s.storage.Wallets(tx...).UpdateUserIPAddress(ctx, repo_wallets.UpdateUserIPAddressParams{
			IpAddress: params.IpAddress,
			ID:        wallet.ID,
		}); err != nil {
			return err
		}
	}

	if params.Locale != "" && wallet.Locale != params.Locale {
		if err := s.UpdateLocale(ctx, wallet.ID, params.Locale, tx...); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) generateWalletAddresses(ctx context.Context, tx pgx.Tx, owner *models.User, wallet *models.Wallet, str *models.Store, currencies []*models.Currency, amount string) ([]*models.WalletAddress, error) {
	currencyIDs := make([]string, 0, len(currencies))
	for _, c := range currencies {
		if !c.IsFiat {
			currencyIDs = append(currencyIDs, c.ID)
		}
	}

	cleanAddresses, err := s.storage.WalletAddresses(repos.WithTx(tx)).GetAllClearByWalletID(ctx, wallet.ID, currencyIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet addresses: %w", err)
	}

	result := make([]*models.WalletAddress, 0, len(currencies))
	for _, c := range currencies {
		if c.IsFiat {
			continue
		}

		idx := slices.IndexFunc(cleanAddresses, func(wa *models.WalletAddress) bool {
			return wa.CurrencyID == c.ID
		})

		var addr *models.WalletAddress
		if idx >= 0 {
			addr = cleanAddresses[idx]
			if logErr := s.logProcessingAddressReceived(ctx, addr, pgtypeutils.DecodeText(wallet.IpAddress)); logErr != nil {
				s.logger.Errorw("failed create log to process processing addresses", "error", logErr)
			}
		} else {
			addr, err = s.getOrCreateWalletAddress(ctx, tx, owner, wallet, c)
			if err != nil {
				return nil, fmt.Errorf("failed to get or create wallet address: %w", err)
			}
		}

		amt, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
			Source:     str.RateSource.String(),
			From:       models.CurrencyCodeUSDT,
			To:         c.Code,
			Amount:     amount,
			StableCoin: c.IsStablecoin,
			Scale:      &str.RateScale,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to convert rate source: %w", err)
		}

		addr.Amount = amt
		result = append(result, addr)
	}

	return result, nil
}
