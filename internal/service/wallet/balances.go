package wallet

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dv-net/dv-merchant/internal/delivery/http/request/wallet_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

type IWalletBalances interface {
	GetWalletBalance(ctx context.Context, dto wallet_request.GetWalletByStoreRequest, rates *exrate.Rates) (*storecmn.FindResponseWithFullPagination[*models.WalletWithUSDBalance], error)
	GetHotWalletsTotalBalance(ctx context.Context, user *models.User) (*AddressesTotalBalance, error)
	GetColdWalletsTotalBalance(ctx context.Context, user *models.User) (*AddressesTotalBalance, error)
	GetExchangeWalletsTotalBalance(ctx context.Context, user *models.User) (*AddressesTotalBalance, error)
	GetProcessingBalances(ctx context.Context, dto GetProcessingWalletsDTO) ([]*ProcessingWalletWithAssets, error)
	ProcessingBalanceStatsInBackground(ctx context.Context, updateInterval time.Duration)
}

// GetWalletBalance TODO change request to dto
func (s *Service) GetWalletBalance(ctx context.Context, dto wallet_request.GetWalletByStoreRequest, rates *exrate.Rates) (*storecmn.FindResponseWithFullPagination[*models.WalletWithUSDBalance], error) {
	commonParams := storecmn.NewCommonFindParams()

	if dto.PageSize != nil {
		commonParams.SetPageSize(dto.PageSize)
	}
	if dto.Page != nil {
		commonParams.SetPage(dto.Page)
	}

	if dto.IsSortByAmount {
		commonParams.OrderBy = "amount"
	}

	if dto.IsSortByBalance {
		commonParams.OrderBy = "amount_usd"
	}

	commonParams.IsAscOrdering = !dto.IsSortByBalance

	params := repo_wallet_addresses.FindParams{
		StoreIDs:         dto.StoreIDs,
		CommonFindParams: *commonParams,
		CurrencyID:       dto.CurrencyID,
		Amount:           dto.Amount,
		Blockchain:       dto.Blockchain,
		Address:          dto.Address,
		WalletIDs:        dto.WalletIDs,
		Rates:            rates.Rate,
		IDs:              rates.CurrencyIDs,
		SortByBalance:    dto.IsSortByBalance,
		SortByAmount:     dto.IsSortByAmount,
		BalanceFrom:      dto.BalanceFiatFrom,
		BalanceTo:        dto.BalanceFiatTo,
	}

	wallets, err := s.storage.WalletAddresses().Find(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = fmt.Errorf("no wallets found by store: %w", err)
		}
		return nil, err
	}
	addresses := make([]*models.WalletWithUSDBalance, 0, len(wallets.Items))

	for _, w := range wallets.Items {
		addresses = append(addresses, &models.WalletWithUSDBalance{
			WalletAddressID: w.ID,
			CurrencyID:      w.CurrencyID,
			Address:         w.Address,
			Blockchain:      w.Blockchain,
			Amount:          w.Amount,
			AmountUSD:       w.AmountUSD,
		})
	}

	return &storecmn.FindResponseWithFullPagination[*models.WalletWithUSDBalance]{
		Items:      addresses,
		Pagination: wallets.Pagination,
	}, nil
}

func (s *Service) GetHotWalletsTotalBalance(ctx context.Context, user *models.User) (*AddressesTotalBalance, error) {
	var total, dust decimal.Decimal

	addresses, err := s.storage.WalletAddresses().GetWalletAddressesTotalWithCurrencyID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all hot wallets: %w", err)
	}

	for _, address := range addresses {
		amountUSD, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
			Source:     user.RateSource.String(),
			From:       address.Code.String,
			To:         models.CurrencyCodeUSD,
			Amount:     address.Balance.String(),
			StableCoin: false,
		})
		if err != nil {
			return nil, fmt.Errorf("error convert amount wallet: %w", err)
		}
		if amountUSD.LessThan(decimal.NewFromInt(1)) {
			dust = dust.Add(amountUSD)
		}
		total = total.Add(amountUSD)
	}

	return &AddressesTotalBalance{
		TotalUSD:  total,
		TotalDust: dust,
	}, nil
}

// GetColdWalletsTotalBalance returns accumulated cold wallets balance
func (s *Service) GetColdWalletsTotalBalance(ctx context.Context, user *models.User) (*AddressesTotalBalance, error) {
	coldAddresses, err := s.storage.WithdrawalWalletAddresses().GetAddressWithCurrencyByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all cold wallets: %w", err)
	}

	preparedAddresses := make([]CalcBalanceDTO, 0, len(coldAddresses))
	for _, address := range coldAddresses {
		preparedAddresses = append(preparedAddresses, CalcBalanceDTO{
			Address:    address.Address,
			Blockchain: address.Blockchain,
		})
	}

	return s.getWalletsTotalBalance(ctx, user, preparedAddresses)
}

// GetExchangeWalletsTotalBalance returns accumulated exchange withdrawal setting wallets balance
func (s *Service) GetExchangeWalletsTotalBalance(ctx context.Context, user *models.User) (*AddressesTotalBalance, error) {
	exchangeAddresses, err := s.storage.ExchangeWithdrawalSettings().GetAllAddressesWithEnabledCurr(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all exchange wallets: %w", err)
	}

	preparedAddresses := make([]CalcBalanceDTO, 0, len(exchangeAddresses))
	for _, address := range exchangeAddresses {
		preparedAddresses = append(preparedAddresses, CalcBalanceDTO{
			Address:    address.Address,
			Blockchain: address.Blockchain,
		})
	}

	return s.getWalletsTotalBalance(ctx, user, preparedAddresses)
}

func (s *Service) GetProcessingBalances(ctx context.Context, dto GetProcessingWalletsDTO) ([]*ProcessingWalletWithAssets, error) {
	enabledCurrencies, err := s.currencyService.GetCurrenciesEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled currencies: %w", err)
	}

	// Sanitize request to ensure valid blockchains and currencies
	{
		// If request is empty, fill with all enabled blockchains and currencies
		if len(dto.Currencies) == 0 && len(dto.Blockchains) == 0 {
			dto.Blockchains = lo.FilterMap(enabledCurrencies, func(c *models.Currency, _ int) (models.Blockchain, bool) {
				if !c.IsFiat && c.Blockchain != nil {
					return *c.Blockchain, true
				}
				return "", false
			})
			dto.Currencies = lo.FilterMap(enabledCurrencies, func(c *models.Currency, _ int) (string, bool) {
				if !c.IsFiat {
					return c.ID, true
				}
				return "", false
			})
		}
		// If we have currencies, ensure they are valid and derive blockchains from them if needed
		{
			if len(dto.Currencies) > 0 { //nolint:nestif
				for _, currID := range dto.Currencies {
					currency, exists := lo.Find(enabledCurrencies, func(c *models.Currency) bool {
						return c.ID == currID
					})
					if !exists {
						return nil, fmt.Errorf("currency %s is disabled", currID)
					}
					if !lo.Contains(dto.Blockchains, *currency.Blockchain) {
						dto.Blockchains = append(dto.Blockchains, *currency.Blockchain)
					}
				}
			} else if len(dto.Blockchains) > 0 {
				for _, blockchain := range dto.Blockchains {
					currencies := lo.FilterMap(enabledCurrencies, func(c *models.Currency, _ int) (string, bool) {
						if c.Blockchain != nil {
							if *c.Blockchain == blockchain && !c.IsFiat {
								return c.ID, true
							}
						}
						return "", false
					})
					dto.Currencies = append(dto.Currencies, currencies...)
				}
			}
		}
	}

	var mu sync.Mutex
	results := make([]*ProcessingWalletWithAssets, 0, len(dto.Blockchains))

	var wg sync.WaitGroup
	for _, blockchain := range lo.Uniq(dto.Blockchains) {
		wg.Add(1)
		go func(blockchain models.Blockchain) {
			defer wg.Done()
			wallets, err := s.processBlockchainWallets(ctx, blockchain, dto, enabledCurrencies)
			if err != nil {
				s.logger.Error("failed to process blockchain wallets", err, "blockchain", blockchain.String())
				return
			}
			if len(wallets) != 0 {
				mu.Lock()
				results = append(results, wallets...)
				mu.Unlock()
			}
		}(blockchain)
	}
	wg.Wait()

	sortWallets(results)
	return results, nil
}
