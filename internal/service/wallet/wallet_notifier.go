package wallet

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

func (s *Service) SendUserWalletNotification(ctx context.Context, walletID uuid.UUID, selectCurrency *string) error {
	walletData, err := s.storage.Wallets().GetFullDataByID(ctx, walletID)
	if err != nil {
		return fmt.Errorf("failed to get full data by id: %w", err)
	}

	availableCurrencies, err := s.storage.StoreCurrencies().GetAllByStoreID(ctx, walletData.Store.ID)
	if err != nil {
		return fmt.Errorf("failed to get available currencies by store id: %w", err)
	}

	availableCurrenciesIDs := make([]string, 0, len(availableCurrencies))
	for _, c := range availableCurrencies {
		availableCurrenciesIDs = append(availableCurrenciesIDs, c.ID)
	}

	addresses, err := s.storage.WalletAddresses().GetAllClearByWalletID(ctx, walletID, availableCurrenciesIDs)
	if err != nil {
		return fmt.Errorf("failed to get all clear addresses by wallet id: %w", err)
	}

	storeOwner, err := s.storage.Users().GetByID(ctx, walletData.Store.UserID)
	if err != nil {
		return fmt.Errorf("failed to get store by wallet id: %w", err)
	}

	hash := s.calculateAddressHash(addresses)

	var targetEmail *string
	if walletData.Wallet.UntrustedEmail.Valid && walletData.Wallet.UntrustedEmail.String != "" {
		targetEmail = &walletData.Wallet.UntrustedEmail.String
	}
	if walletData.Wallet.Email.Valid && walletData.Wallet.Email.String != "" {
		targetEmail = &walletData.Wallet.Email.String
	}

	if targetEmail != nil {
		s.notifyStoreOwnerWalletsList(ctx, notifyStoreOwnerWalletsListParams{
			User:            storeOwner,
			StoreID:         walletData.Store.ID,
			WalletAddresses: addresses,
			Hash:            hash,
			WalletEmail:     *targetEmail,
			Locale:          &walletData.Wallet.Locale,
			SelectCurrency:  selectCurrency,
		})
	}
	return nil
}

type notifyStoreOwnerWalletsListParams struct {
	User            *models.User
	StoreID         uuid.UUID
	WalletAddresses []*models.WalletAddress
	Hash            string
	WalletEmail     string
	Locale          *string
	SelectCurrency  *string
}

func (s *Service) notifyStoreOwnerWalletsList(ctx context.Context, params notifyStoreOwnerWalletsListParams) {
	sort.Slice(params.WalletAddresses, func(i, j int) bool {
		walletI := params.WalletAddresses[i]
		walletJ := params.WalletAddresses[j]

		if params.SelectCurrency != nil {
			if walletI.CurrencyID == *params.SelectCurrency && walletJ.CurrencyID != *params.SelectCurrency {
				return true
			}
			if walletI.CurrencyID != *params.SelectCurrency && walletJ.CurrencyID == *params.SelectCurrency {
				return false
			}
		}

		blockchainI := walletI.Blockchain
		blockchainJ := walletJ.Blockchain
		if blockchainI != blockchainJ {
			return string(blockchainI) < string(blockchainJ)
		}

		nativeCurrency, _ := blockchainI.NativeCurrency()
		isNativeI := nativeCurrency == walletI.CurrencyID
		isNativeJ := nativeCurrency == walletJ.CurrencyID

		if isNativeI == isNativeJ {
			return false
		}
		return isNativeI
	})

	notificationWalletsData := make([]notify.WalletDTO, 0, len(params.WalletAddresses))
	for i, wallet := range params.WalletAddresses {
		walletDTO := notify.WalletDTO{
			CurrencyID:   wallet.CurrencyID,
			CurrencyName: wallet.CurrencyID,
			Address:      wallet.Address,
		}
		nativeCurrency, _ := wallet.Blockchain.NativeCurrency()
		if nativeCurrency != wallet.CurrencyID {
			walletDTO.ShowBlockchain = true
			walletDTO.BlockchainID = wallet.Blockchain.String()
			walletDTO.BlockchainName = strings.ToUpper(wallet.Blockchain.String())
		}
		if i == 0 {
			walletDTO.IsFirst = true
		}
		notificationWalletsData = append(notificationWalletsData, walletDTO)
	}

	language := params.User.Language
	if params.Locale != nil && *params.Locale != "" {
		language = util.ParseLanguageTag(*params.Locale).String()
	}
	if language == "" {
		language = util.ParseLanguageTag("").String()
	}

	payload := &notify.ExternalWalletRequestedData{
		Language:         language,
		Addresses:        notificationWalletsData,
		NotificationHash: params.Hash,
	}

	s.logger.Debugw("External wallets request payload", "payload", payload)

	go s.notification.SendSystemEmail(ctx, models.NotificationTypeExternalWalletRequested, params.WalletEmail, payload, &models.NotificationArgs{UserID: &params.User.ID, StoreID: &params.StoreID})
}

func (s *Service) calculateAddressHash(addresses []*models.WalletAddress) string {
	hashes := lo.Map(addresses, func(w *models.WalletAddress, _ int) string {
		h := sha256.New()
		if _, err := h.Write([]byte(w.Address)); err != nil {
			s.logger.Errorw("failed to hash address", "error", err)
			return ""
		}
		return hex.EncodeToString(h.Sum(nil))[:6]
	})
	return strings.Join(lo.Uniq(hashes), "-")
}
