package wallet

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"connectrpc.com/connect"
	addressesv2 "github.com/dv-net/dv-proto/gen/go/eproxy/addresses/v2"
	evmv2 "github.com/dv-net/dv-proto/gen/go/eproxy/evm/v2"

	"github.com/dv-net/dv-merchant/internal/config"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/wallet_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/currency"
	"github.com/dv-net/dv-merchant/internal/service/eproxy"
	"github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/internal/service/notify"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallets"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/dv-merchant/pkg/pgtypeutils"

	"github.com/gocarina/gocsv"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/puzpuzpuz/xsync/v3"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

const (
	TRC20EnergyPriceTRX    = 28
	TRC20BandwidthPriceTRX = 1
	TRXBandwidthPriceTRX   = 0.348
)

// ChainConfig holds chain-specific gas parameters
type ChainConfig struct {
	NativeTransferGas   int64   // Gas limit for native token transfer
	ERC20TransferGas    int64   // Gas limit for ERC20 token transfer
	IsL2                bool    // Is this an L2 chain requiring L1 data fees
	L1DataFeeMultiplier float64 // Rough multiplier for L1 data fees
}

// Common chain configurations with typical gas limits
var ChainConfigs = map[models.Blockchain]ChainConfig{
	models.BlockchainEthereum: {
		NativeTransferGas: 21000,
		ERC20TransferGas:  65000, // Average: simple transfers ~46k, complex ones up to 100k
		IsL2:              false,
	},
	models.BlockchainArbitrum: {
		NativeTransferGas:   21000,
		ERC20TransferGas:    80000, // L2 overhead
		IsL2:                true,
		L1DataFeeMultiplier: 1.5, // Arbitrum L1 fees typically 1.5x the L2 execution fee
	},
	models.BlockchainBinanceSmartChain: {
		NativeTransferGas: 21000,
		ERC20TransferGas:  60000,
		IsL2:              false,
	},
	models.BlockchainPolygon: {
		NativeTransferGas: 21000,
		ERC20TransferGas:  65000,
		IsL2:              false, // Polygon is a sidechain, not L2
	},
}

type IWalletService interface {
	GetWallet(ctx context.Context, ID uuid.UUID) (*models.Wallet, error)
	StoreWalletWithAddress(ctx context.Context, dto CreateStoreWalletWithAddressDTO, amount string) (*WithAddressDto, error)
	GetFullDataByID(ctx context.Context, ID uuid.UUID) (*GetAllByStoreIDResponse, error)
	GetProcessingBalances(ctx context.Context, dto GetProcessingWalletsDTO) ([]*ProcessingWalletWithAssets, error)
	SummarizeUserWalletsByCurrency(ctx context.Context, userID uuid.UUID, rates *exrate.Rates, minBalance decimal.Decimal) ([]SummaryDTO, error)
	GetWalletsInfo(ctx context.Context, userID uuid.UUID, address string) ([]*WithBlockchains, error)
	LoadPrivateAddresses(ctx context.Context, dto LoadPrivateKeyDTO) (*bytes.Buffer, error)
	ProcessingBalanceStatsInBackground(ctx context.Context, updateInterval time.Duration)
	FetchTronResourceStatistics(ctx context.Context, user *models.User, dto FetchTronStatisticsParams) (map[string]CombinedStats, error)
	UpdateLocale(ctx context.Context, walletID uuid.UUID, locale string) error
}

type IWalletBalances interface {
	GetWalletBalance(ctx context.Context, dto wallet_request.GetWalletByStoreRequest, rates *exrate.Rates) (*storecmn.FindResponseWithFullPagination[*models.WalletWithUSDBalance], error)
	GetHotWalletsTotalBalance(ctx context.Context, user *models.User) (*AddressesTotalBalance, error)
	GetColdWalletsTotalBalance(ctx context.Context, user *models.User) (*AddressesTotalBalance, error)
	GetExchangeWalletsTotalBalance(ctx context.Context, user *models.User) (*AddressesTotalBalance, error)
}

type Service struct {
	cfg               *config.Config
	storage           storage.IStorage
	logger            logger.Logger
	currencyService   currency.ICurrency
	processingService processing.IProcessingWallet
	exrateService     exrate.IExRateSource
	currConvService   currconv.ICurrencyConvertor
	settingService    setting.ISettingService
	eproxyService     eproxy.IExplorerProxy
	notification      notify.INotificationService

	updateBalanceInProgress         atomic.Bool
	updateProcessingStatsInProgress atomic.Bool
}

var _ IWalletService = (*Service)(nil)
var _ IWalletBalances = (*Service)(nil)

func New(
	cfg *config.Config,
	storage storage.IStorage,
	logger logger.Logger,
	currencyService currency.ICurrency,
	processingService processing.IProcessingWallet,
	exrateService exrate.IExRateSource,
	currConvService currconv.ICurrencyConvertor,
	settingsService setting.ISettingService,
	eproxyService eproxy.IExplorerProxy,
	notification notify.INotificationService,
) *Service {
	return &Service{
		cfg:               cfg,
		storage:           storage,
		logger:            logger,
		currencyService:   currencyService,
		processingService: processingService,
		exrateService:     exrateService,
		currConvService:   currConvService,
		settingService:    settingsService,
		eproxyService:     eproxyService,
		notification:      notification,
	}
}

func (s *Service) GetWallet(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	wallet, err := s.storage.Wallets().GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *Service) UpdateLocale(ctx context.Context, walletID uuid.UUID, locale string) error {
	return s.storage.Wallets().UpdateUserLocale(ctx, repo_wallets.UpdateUserLocaleParams{
		Locale: locale,
		ID:     walletID,
	})
}

// GetFullDataByID returns wallet with store, addresses and available currencies
func (s *Service) GetFullDataByID(ctx context.Context, id uuid.UUID) (*GetAllByStoreIDResponse, error) {
	// get wallet data
	data, err := s.storage.Wallets().GetFullDataByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get full data by id: %w", err)
	}

	// get all available currencies by store id
	availableCurrencies, err := s.storage.StoreCurrencies().GetAllByStoreID(ctx, data.Store.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available currencies by store id: %w", err)
	}

	availableCurrenciesIDs := make([]string, 0, len(availableCurrencies))
	for _, c := range availableCurrencies {
		availableCurrenciesIDs = append(availableCurrenciesIDs, c.ID)
	}

	// get rates data
	rates, err := s.exrateService.GetStoreCurrencyRate(ctx, availableCurrencies, data.Store.RateSource.String(), data.Store.RateScale)
	if err != nil {
		return nil, fmt.Errorf("failed to get store currency rate: %w", err)
	}

	// get all clear addresses by wallet id
	addresses, err := s.storage.WalletAddresses().GetAllClearByWalletID(ctx, id, availableCurrenciesIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get all clear addresses by wallet id: %w", err)
	}

	var storeOwner *models.User
	for _, c := range availableCurrencies {
		if c.IsFiat {
			continue
		}

		exists := slices.ContainsFunc(addresses, func(wa *models.WalletAddress) bool {
			return wa.CurrencyID == c.ID
		})

		if exists {
			continue
		}

		/*
			Create address in processing service
		*/

		// get store owner once if needed
		if storeOwner == nil {
			storeOwner, err = s.storage.Users().GetByID(ctx, data.Store.UserID)
			if err != nil {
				return nil, fmt.Errorf("failed to get store owner by id: %w", err)
			}
		}

		// get or create wallet address
		newWalletAddress, err := s.getOrCreateWalletAddress(ctx, nil, storeOwner, &data.Wallet, c)
		if err != nil {
			return nil, fmt.Errorf("failed to get or create wallet address: %w", err)
		}

		addresses = append(addresses, newWalletAddress)
	}

	result := &GetAllByStoreIDResponse{
		GetFullDataByIDRow:  *data,
		Addresses:           addresses,
		AvailableCurrencies: availableCurrencies,
		Rates:               rates,
	}

	return result, nil
}

// getOrCreateWalletAddress returns wallet with addresses
func (s *Service) getOrCreateWalletAddress(
	ctx context.Context,
	dbTx pgx.Tx,
	storeOwner *models.User,
	wallet *models.Wallet,
	c *models.Currency,
) (*models.WalletAddress, error) {
	if c.IsFiat {
		return nil, fmt.Errorf("failed to create address for fiat currency")
	}

	if c.Blockchain == nil || *c.Blockchain == "" {
		return nil, fmt.Errorf("blockchain is not set for currency %s", c.ID)
	}

	walletAddress, err := s.storage.WalletAddresses().GetByWalletIDAndCurrencyID(ctx, wallet.ID, c.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to get wallet address: %w", err)
	}

	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return s.createNewWalletAddress(ctx, dbTx, storeOwner, wallet, c, nil)
	}

	if err == nil && walletAddress.Dirty {
		return s.createNewWalletAddress(ctx, dbTx, storeOwner, wallet, c, walletAddress)
	}

	err = s.logProcessingAddressReceived(ctx, walletAddress, pgtypeutils.DecodeText(wallet.IpAddress))
	if err != nil {
		s.logger.Error("failed create log to process processing addresses", err)
	}
	return walletAddress, nil
}

func (s *Service) createNewWalletAddress(
	ctx context.Context,
	dbTx pgx.Tx,
	storeOwner *models.User,
	wallet *models.Wallet,
	c *models.Currency,
	oldWalletAddress *models.WalletAddress,
) (*models.WalletAddress, error) {
	if oldWalletAddress != nil {
		if err := s.processingService.MarkDirtyHotWallet(ctx, storeOwner.ProcessingOwnerID.UUID, *c.Blockchain, oldWalletAddress.Address); err != nil {
			return nil, fmt.Errorf("failed to mark dirty hot wallet: %w", err)
		}
	}

	params := processing.CreateOwnerHotWalletParams{
		OwnerID:    storeOwner.ProcessingOwnerID.UUID,
		CustomerID: wallet.ID.String(),
		Blockchain: *c.Blockchain,
	}

	switch *c.Blockchain {
	case models.BlockchainBitcoin:
		params.BitcoinAddressType = util.Pointer(processing.ConvertToBitcoinAddressType(s.cfg.Blockchain.Bitcoin.AddressType))
	case models.BlockchainLitecoin:
		params.LitecoinAddressType = util.Pointer(processing.ConvertToLitecoinAddressType(s.cfg.Blockchain.Litecoin.AddressType))
	}

	newWallet, err := s.processingService.CreateOwnerHotWallet(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create new hot wallet: %w", err)
	}

	if oldWalletAddress != nil && newWallet.Address == oldWalletAddress.Address {
		return nil, fmt.Errorf("failed to create new wallet address: new address is the same as the old one")
	}

	walletAddress, err := s.storage.WalletAddresses(repos.WithTx(dbTx)).Create(ctx, repo_wallet_addresses.CreateParams{
		WalletID:   wallet.ID,
		UserID:     storeOwner.ID,
		CurrencyID: c.ID,
		Blockchain: *c.Blockchain,
		Address:    newWallet.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create new wallet address: %w", err)
	}

	return walletAddress, nil
}

// StoreWalletWithAddress creates/returns wallet with addresses
func (s *Service) StoreWalletWithAddress(ctx context.Context, dto CreateStoreWalletWithAddressDTO, amountUSD string) (*WithAddressDto, error) {
	var storeOwner *models.User
	var storeID uuid.UUID
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

		wallet.ID = w.ID
		err = s.updateWalletMeta(ctx, wallet, dto.ToCreateParams(), &walletEmail, repos.WithTx(tx))
		if err != nil {
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
		storeID = str.ID

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

	hash := s.calculateAddressHash(walletWithAddress.Address)

	s.logger.Info("wallet with address created", "store_external_id", dto.StoreExternalID, "hash", hash)

	var targetEmail *string
	if walletEmail != nil && *walletEmail != "" {
		targetEmail = walletEmail
	}
	if dto.UntrustedEmail != nil && *dto.UntrustedEmail != "" {
		targetEmail = dto.UntrustedEmail
	}
	if targetEmail != nil {
		s.notifyStoreOwnerWalletsList(ctx, storeOwner, storeID, walletWithAddress.Address, hash, *targetEmail, dto.Locale)
	}

	return walletWithAddress, nil
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
			err = errors.Wrap(err, "no wallets found by store")
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

func (s *Service) processBlockchainWallets(ctx context.Context, blockchain models.Blockchain, dto GetProcessingWalletsDTO, enabledCurrencies []*models.Currency) ([]*ProcessingWalletWithAssets, error) {
	wallets, err := s.processingService.GetOwnerProcessingWallets(ctx, processing.GetOwnerProcessingWalletsParams{
		OwnerID:    dto.OwnerID,
		Blockchain: util.Pointer(blockchain),
	})
	if err != nil {
		if strings.Contains(err.Error(), "not available") {
			return nil, nil // Not an error, just not available
		}
		return nil, fmt.Errorf("failed to fetch wallets: %w", err)
	}

	result := make([]*ProcessingWalletWithAssets, 0, len(wallets))
	for _, wallet := range wallets {
		processingWallet, err := s.buildProcessingWallet(ctx, wallet, blockchain, dto.Currencies, enabledCurrencies)
		if err != nil {
			return nil, fmt.Errorf("failed to build processing wallet: %w", err)
		}
		result = append(result, processingWallet)
	}

	return result, nil
}

func (s *Service) buildProcessingWallet(ctx context.Context, wallet processing.WalletProcessing, blockchain models.Blockchain, currencies []string, enabledCurrencies []*models.Currency) (*ProcessingWalletWithAssets, error) {
	nativeCurrency, err := blockchain.NativeCurrency()
	if err != nil {
		return nil, fmt.Errorf("failed to get native currency for blockchain %s: %w", blockchain, err)
	}

	currency, exists := lo.Find(enabledCurrencies, func(c *models.Currency) bool {
		return c.ID == nativeCurrency && c.Blockchain != nil && *c.Blockchain == blockchain
	})
	if !exists {
		return nil, fmt.Errorf("native currency %s for blockchain %s is not enabled", nativeCurrency, blockchain)
	}

	processingWallet := &ProcessingWalletWithAssets{
		Address:    wallet.Address,
		Blockchain: wallet.Blockchain,
		Currency: &models.CurrencyShort{
			ID:            currency.ID,
			Code:          currency.Code,
			Precision:     currency.Precision,
			Name:          currency.Name,
			Blockchain:    currency.Blockchain,
			IsEVMLike:     currency.Blockchain.IsEVMLike(),
			IsBitcoinLike: currency.Blockchain.IsBitcoinLike(),
		},
	}

	// Assemble assets and balance
	if err := s.assembleWalletBalance(ctx, processingWallet, wallet, blockchain, currencies, enabledCurrencies); err != nil {
		return nil, fmt.Errorf("failed to assemble wallet balance: %w", err)
	}

	// Add blockchain-specific additional data
	if lo.Contains(currencies, nativeCurrency) {
		if err := s.addBlockchainSpecificData(ctx, processingWallet, wallet, blockchain); err != nil {
			return nil, fmt.Errorf("failed to add blockchain specific data: %w", err)
		}
	}

	return processingWallet, nil
}

func (s *Service) assembleWalletBalance(ctx context.Context, processingWallet *ProcessingWalletWithAssets, wallet processing.WalletProcessing, blockchain models.Blockchain, currencies []string, enabledCurrencies []*models.Currency) error {
	// Native currency
	nativeCurrency, err := blockchain.NativeCurrency()
	if err != nil {
		return fmt.Errorf("failed to get native currency for blockchain %s: %w", blockchain, err)
	}
	// Assemble assets
	assets, err := s.assembleAssets(ctx, blockchain, currencies, enabledCurrencies, wallet.Assets)
	if err != nil {
		return fmt.Errorf("failed to assemble assets: %w", err)
	}
	processingWallet.Assets = assets

	// Find and set native token balance
	asset, exists := lo.Find(assets, func(a *Asset) bool {
		return nativeCurrency == a.CurrencyID
	})
	if exists {
		processingWallet.Balance = &Balance{
			NativeToken:    asset.Amount,
			NativeTokenUSD: asset.AmountUSD,
		}
	}

	return nil
}

func (s *Service) addBlockchainSpecificData(ctx context.Context, processingWallet *ProcessingWalletWithAssets, wallet processing.WalletProcessing, blockchain models.Blockchain) error {
	switch blockchain {
	case models.BlockchainTron:
		return s.addTronData(processingWallet, wallet)
	default:
		if blockchain.IsEVMLike() {
			return s.addEVMData(ctx, processingWallet, blockchain)
		}
	}
	return nil
}

func (s *Service) addTronData(processingWallet *ProcessingWalletWithAssets, wallet processing.WalletProcessing) error {
	if wallet.AdditionalData == nil || wallet.AdditionalData.TronData == nil {
		return nil
	}

	tronData := &TronData{
		AvailableEnergyForUse:    wallet.AdditionalData.TronData.AvailableEnergyForUse,
		TotalEnergy:              wallet.AdditionalData.TronData.TotalEnergy,
		AvailableBandwidthForUse: wallet.AdditionalData.TronData.AvailableBandwidthForUse,
		TotalBandwidth:           wallet.AdditionalData.TronData.TotalBandwidth,
		StackedTrx:               wallet.AdditionalData.TronData.StackedTrx,
		StackedEnergy:            wallet.AdditionalData.TronData.StackedEnergy,
		StackedEnergyTrx:         wallet.AdditionalData.TronData.StackedEnergyTrx,
		StackedBandwidth:         wallet.AdditionalData.TronData.StackedBandwidth,
		StackedBandwidthTrx:      wallet.AdditionalData.TronData.StackedBandwidthTrx,
		TotalUsedEnergy:          wallet.AdditionalData.TronData.TotalUsedEnergy,
		TotalUsedBandwidth:       wallet.AdditionalData.TronData.TotalUsedBandwidth,
	}

	// Calculate max transfers if balance is available
	if processingWallet.Balance != nil {
		if err := s.calculateTronTransfers(tronData, processingWallet.Balance.NativeToken); err != nil {
			return fmt.Errorf("failed to calculate tron transfers: %w", err)
		}
	}

	processingWallet.AdditionalData = &BlockchainAdditionalData{TronData: tronData}
	return nil
}

func (s *Service) calculateTronTransfers(tronData *TronData, nativeTokenBalance string) error {
	tokenBalance, err := decimal.NewFromString(nativeTokenBalance)
	if err != nil {
		return fmt.Errorf("failed to parse token balance: %w", err)
	}

	maxTransfersNative := tokenBalance.Div(decimal.NewFromFloat(TRXBandwidthPriceTRX)).Floor().String()
	maxTransfersTRC20 := tokenBalance.Div(decimal.NewFromFloat(TRC20EnergyPriceTRX).Add(decimal.NewFromFloat(TRC20BandwidthPriceTRX))).Floor().String()

	tronData.MaxTransfersNative = maxTransfersNative
	tronData.MaxTransfersTRC20 = maxTransfersTRC20

	return nil
}

func (s *Service) addEVMData(ctx context.Context, processingWallet *ProcessingWalletWithAssets, blockchain models.Blockchain) error {
	eBlockchain, err := blockchain.ToEPb()
	if err != nil {
		return fmt.Errorf("failed to convert blockchain [%s] to pb: %w", blockchain, err)
	}

	estimatedGasPrice, err := s.eproxyService.EVM().SuggestGasPrice(ctx, connect.NewRequest(&evmv2.SuggestGasPriceRequest{
		Blockchain: eBlockchain,
	}))
	if err != nil {
		return fmt.Errorf("failed to get gas price: %w", err)
	}

	evmData := &EVMData{
		SuggestedGasPrice: estimatedGasPrice.Msg.GasFeeWei,
	}

	// Calculate max transfers if balance is available
	if processingWallet.Balance != nil {
		if err := s.calculateEVMTransfers(evmData, processingWallet.Balance.NativeToken, estimatedGasPrice.Msg.GasFeeWei, blockchain); err != nil {
			return fmt.Errorf("failed to calculate EVM transfers: %w", err)
		}
	}

	processingWallet.AdditionalData = &BlockchainAdditionalData{EVMData: evmData}
	return nil
}

func (s *Service) calculateEVMTransfers(evmData *EVMData, nativeTokenBalance, gasFeeWei string, blockchain models.Blockchain) error {
	chainConfig, exists := ChainConfigs[blockchain]
	if !exists {
		return fmt.Errorf("unsupported chain: %s", blockchain)
	}

	gasPrice, err := decimal.NewFromString(gasFeeWei)
	if err != nil {
		return fmt.Errorf("failed to parse gas price: %w", err)
	}

	gasPriceEth := gasPrice.Div(decimal.NewFromInt(1e18))

	balance, err := decimal.NewFromString(nativeTokenBalance)
	if err != nil {
		return fmt.Errorf("failed to parse balance: %w", err)
	}

	// Calculate L2 execution cost
	nativeTransferCost := gasPriceEth.Mul(decimal.NewFromInt(chainConfig.NativeTransferGas))
	erc20TransferCost := gasPriceEth.Mul(decimal.NewFromInt(chainConfig.ERC20TransferGas))

	// For L2 chains, add estimated L1 data fees
	if chainConfig.IsL2 {
		// L1 data fee estimation based on multiplier
		l1DataFeeNative := nativeTransferCost.Mul(decimal.NewFromFloat(chainConfig.L1DataFeeMultiplier))
		l1DataFeeERC20 := erc20TransferCost.Mul(decimal.NewFromFloat(chainConfig.L1DataFeeMultiplier))

		// Total cost = L2 execution + L1 data fee
		nativeTransferCost = nativeTransferCost.Add(l1DataFeeNative)
		erc20TransferCost = erc20TransferCost.Add(l1DataFeeERC20)

		evmData.L1DataFeeEstimate = l1DataFeeNative.String()
	}

	maxNativeTransfers := balance.Div(nativeTransferCost).IntPart()
	maxERC20Transfers := balance.Div(erc20TransferCost).IntPart()

	evmData.IsL2 = chainConfig.IsL2
	evmData.MaxTransfersNative = strconv.FormatInt(maxNativeTransfers, 10)
	evmData.MaxTransfersERC20 = strconv.FormatInt(maxERC20Transfers, 10)
	evmData.CostPerNative = nativeTransferCost.String()
	evmData.CostPerERC20 = erc20TransferCost.String()

	return nil
}

func (s *Service) assembleAssets(ctx context.Context, blockchain models.Blockchain, filterCurrencies []string, enabledCurrencies []*models.Currency, processingAssets []*processing.Asset) ([]*Asset, error) {
	currencies, err := s.currencyService.GetCurrenciesByBlockchain(ctx, blockchain)
	if err != nil {
		return nil, fmt.Errorf("failed to get currencies by blockchain: %w", err)
	}

	filterCurrencies = lo.FilterMap(filterCurrencies, func(c string, _ int) (string, bool) {
		for _, currency := range currencies {
			if c == currency.ID {
				return c, true
			}
		}
		return "", false
	})

	assets := make([]*Asset, 0, len(filterCurrencies))
	for _, currency := range enabledCurrencies {
		if slices.Contains(filterCurrencies, currency.ID) {
			if asset, exists := lo.Find(processingAssets, func(a *processing.Asset) bool {
				return a.Identity == currency.ID || a.Identity == currency.ContractAddress.String
			}); exists {
				// Calculate USD amount
				amountUSD, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
					Source:     models.ExchangeSlugBinance.String(),
					From:       currency.Code,
					To:         models.CurrencyCodeUSDT,
					Amount:     asset.Amount,
					StableCoin: false,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to convert %s %s to USDT: %w", asset.Amount, asset.Identity, err)
				}
				processingAsset := &Asset{
					CurrencyID: currency.ID,
					Identity:   currency.Code,
					Amount:     asset.Amount,
					AmountUSD:  amountUSD.String(),
				}

				assets = append(assets, processingAsset)
			} else {
				assets = append(assets, &Asset{
					CurrencyID: currency.ID,
					Amount:     "0",
					AmountUSD:  "0",
					Identity:   currency.Code,
				})
			}
		}
	}
	return assets, nil
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

func (s *Service) getWalletsTotalBalance(ctx context.Context, user *models.User, addresses []CalcBalanceDTO) (*AddressesTotalBalance, error) {
	total := xsync.NewMapOf[string, decimal.Decimal]()

	var wg sync.WaitGroup
	errCh := make(chan error, len(addresses))

	for _, wallet := range addresses {
		wg.Add(1)
		go func(w CalcBalanceDTO) {
			defer wg.Done()
			walletBalanceTotal, err := s.processWalletBalance(ctx, user, w)
			if err != nil {
				errCh <- fmt.Errorf("failed to process wallet %s", w.Address)
				return
			}

			if v, ok := total.Load(w.Address); ok {
				total.Store(w.Address, v.Add(walletBalanceTotal))
			} else {
				total.Store(w.Address, walletBalanceTotal)
			}
		}(wallet)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		s.logger.Error("fetch wallet balance", err)
	}

	totalUSD := decimal.Zero
	total.Range(func(_ string, v decimal.Decimal) bool {
		totalUSD = totalUSD.Add(v)
		return true
	})

	return &AddressesTotalBalance{
		TotalUSD: totalUSD,
	}, nil
}

// processWalletBalance calculate all assets amount in USD by wallet
func (s *Service) processWalletBalance(
	ctx context.Context,
	user *models.User,
	wallet CalcBalanceDTO,
) (decimal.Decimal, error) {
	blockchain, err := wallet.Blockchain.ToEPb()
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to convert blockchain [%s] to pb: %w", blockchain, err)
	}

	req := &addressesv2.InfoRequest{
		Address:    wallet.Address,
		Blockchain: blockchain,
	}

	walletsRes, err := s.eproxyService.Wallets().Info(ctx, connect.NewRequest(req))
	if err != nil {
		return decimal.Zero, fmt.Errorf("[%s]: failed to get wallet by address: %w", blockchain, err)
	}

	currencies, err := s.currencyService.GetCurrenciesEnabled(ctx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get enabled currencies: %w", err)
	}

	totalAmount := decimal.Zero
	for _, asset := range s.filterProxyEnabledAssets(ctx, currencies, walletsRes.Msg.Item.Assets) {
		amountUSD, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
			Source:     user.RateSource.String(),
			From:       asset.Symbol,
			To:         models.CurrencyCodeUSD,
			Amount:     asset.Amount,
			StableCoin: false,
		})
		if err != nil {
			return decimal.Zero, fmt.Errorf("error converting amount for wallet: %w", err)
		}

		totalAmount = totalAmount.Add(amountUSD)
	}

	return totalAmount, nil
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

func (s *Service) GetWalletsInfo(ctx context.Context, userID uuid.UUID, searchCriteria string) ([]*WithBlockchains, error) {
	walletsData, err := s.storage.Wallets().SearchByParam(ctx, repo_wallets.SearchByParamParams{
		Criteria: pgtype.Text{
			String: searchCriteria,
			Valid:  true,
		},
		UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) || len(walletsData) == 0 {
		return nil, ErrServiceWalletNotFound
	}
	if err != nil {
		return nil, err
	}

	wallets, err := s.groupWalletsData(ctx, walletsData)
	if err != nil {
		return nil, err
	}

	return wallets, nil
}

func (s *Service) groupWalletsData(ctx context.Context, rows []*repo_wallets.SearchByParamRow) ([]*WithBlockchains, error) {
	walletMap := make(map[string]*WithBlockchains)

	for _, row := range rows {
		if _, exists := walletMap[row.Address]; !exists {
			logs, err := s.GetWalletLogs(ctx, row.WalletAddressID)
			if err != nil {
				return nil, fmt.Errorf("error fetching logs for wallet %s: %w", row.WalletID, err)
			}

			wallet := &WithBlockchains{
				WalletCreatedAt: row.WalletCreatedAt.Time,
				WalletID:        row.WalletID,
				StoreExternalID: row.StoreExternalID,
				StoreID:         row.StoreID,
				StoreName:       row.StoreName,
				Address:         row.Address,
				TotalTx:         decimal.Zero,
				Blockchains:     []BlockchainGroup{},
				Logs:            logs,
			}
			if row.Email.Valid {
				email := row.Email.String
				wallet.Email = &email
			}
			walletMap[row.Address] = wallet
		}

		amountUsd, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
			Source:     "binance",
			From:       row.CurrencyCode,
			To:         models.CurrencyCodeUSDT,
			Amount:     row.Amount.String(),
			StableCoin: false,
		})
		if err != nil {
			return nil, fmt.Errorf("error converting amount for wallet %s: %w", row.WalletID, err)
		}

		wallet := walletMap[row.Address]
		asset := AssetWallet{
			Currency:     row.CurrencyID,
			Amount:       row.Amount,
			AmountUSD:    amountUsd,
			TxCount:      row.DepositsCount,
			TotalDeposit: row.DepositsSum,
		}

		found := false
		for i, bg := range wallet.Blockchains {
			if bg.Blockchain == row.Blockchain {
				wallet.Blockchains[i].Assets = append(wallet.Blockchains[i].Assets, asset)
				found = true
				break
			}
		}

		if !found {
			wallet.Blockchains = append(wallet.Blockchains, BlockchainGroup{
				Blockchain: row.Blockchain,
				Assets:     []AssetWallet{asset},
			})
		}
	}

	result := make([]*WithBlockchains, 0, len(walletMap))
	for _, wallet := range walletMap {
		totalTx := lo.Reduce(wallet.Blockchains, func(acc decimal.Decimal, bg BlockchainGroup, _ int) decimal.Decimal {
			return acc.Add(lo.Reduce(bg.Assets, func(acc decimal.Decimal, asset AssetWallet, _ int) decimal.Decimal {
				return acc.Add(asset.TxCount)
			}, decimal.Zero))
		}, decimal.Zero)
		wallet.TotalTx = totalTx
		result = append(result, wallet)
	}

	return result, nil
}

func (s *Service) filterProxyEnabledAssets(_ context.Context, currencies []*models.Currency, assets []*addressesv2.AssetInfo) []*addressesv2.AssetInfo {
	res := make([]*addressesv2.AssetInfo, 0)
	for _, asset := range assets {
		if idx := slices.IndexFunc(currencies, func(c *models.Currency) bool {
			return strings.EqualFold(asset.AssetIdentifier, c.ContractAddress.String)
		}); idx != -1 {
			res = append(res, asset)
		}
	}

	return res
}

func (s *Service) notifyStoreOwnerWalletsList(ctx context.Context, usr *models.User, storeID uuid.UUID, walletAddresses []*models.WalletAddress, hash string, walletEmail string, locale *string) {
	// Sort wallet addresses: group by blockchain, native token first in each group
	sort.Slice(walletAddresses, func(i, j int) bool {
		blockchainI := walletAddresses[i].Blockchain
		blockchainJ := walletAddresses[j].Blockchain

		if blockchainI != blockchainJ {
			return string(blockchainI) < string(blockchainJ)
		}

		nativeCurrency, _ := blockchainI.NativeCurrency()
		isNativeI := nativeCurrency == walletAddresses[i].CurrencyID
		isNativeJ := nativeCurrency == walletAddresses[j].CurrencyID

		if isNativeI == isNativeJ {
			return false
		}
		return isNativeI
	})

	notificationWalletsData := make([]notify.WalletDTO, 0, len(walletAddresses))
	for _, wallet := range walletAddresses {
		walletDTO := notify.WalletDTO{
			CurrencyID:   wallet.CurrencyID,
			CurrencyName: wallet.CurrencyID,
			Address:      wallet.Address,
		}
		// If currency is not native currency to the blockchain, show it
		nativeCurrency, _ := wallet.Blockchain.NativeCurrency()
		if nativeCurrency != wallet.CurrencyID {
			walletDTO.ShowBlockchain = true
			walletDTO.BlockchainID = wallet.Blockchain.String()
			walletDTO.BlockchainName = strings.ToUpper(wallet.Blockchain.String())
		}
		notificationWalletsData = append(notificationWalletsData, walletDTO)
	}

	// Use provided locale from frontend if available, fallback to user's language, then to English as final fallback
	language := usr.Language
	if locale != nil && *locale != "" {
		// Validate and normalize the locale using the existing utility
		language = util.ParseLanguageTag(*locale).String()
	}
	// Ensure we always have a valid language, fallback to English if user language is empty
	if language == "" {
		language = util.ParseLanguageTag("").String() // This will default to English
	}

	payload := &notify.ExternalWalletRequestedData{
		Language:         language,
		Addresses:        notificationWalletsData,
		NotificationHash: hash,
	}

	s.logger.Info("External wallets request payload", "payload", payload)

	go s.notification.SendSystemEmail(ctx, models.NotificationTypeExternalWalletRequested, walletEmail, payload, &models.NotificationArgs{UserID: &usr.ID, StoreID: &storeID})
}

func (s *Service) LoadPrivateAddresses(ctx context.Context, dto LoadPrivateKeyDTO) (*bytes.Buffer, error) {
	data, err := s.processingService.GetOwnerHotWalletKeys(ctx, dto.User, dto.Otp, processing.GetOwnerHotWalletKeysParams{
		WalletAddressIDs:           dto.WalletAddressIDs,
		ExcludedWalletAddressesIDs: dto.ExcludedWalletAddressesIDs,
	})
	if err != nil {
		return nil, err
	}

	bb := new(bytes.Buffer)
	switch dto.FileType {
	case "json":
		j, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		if _, err := bb.Write(j); err != nil {
			return nil, err
		}
	case "csv":
		keys := make([]*HotKeyCsv, 0, len(data.Entries))
		for _, entry := range data.Entries {
			for _, item := range entry.Items {
				keys = append(keys, &HotKeyCsv{
					Blockchain: entry.Name,
					PublicKey:  item.PublicKey,
					PrivateKey: item.PrivateKey,
					Address:    item.Address,
				})
			}
		}
		if err := gocsv.Marshal(keys, bb); err != nil {
			return nil, err
		}
	case "txt":
		for _, entry := range data.Entries {
			for _, item := range entry.Items {
				if _, err := bb.WriteString(item.PrivateKey + "\n"); err != nil {
					return nil, err
				}
			}
		}
	}

	for _, item := range data.AllSelectedWallets {
		err := s.logProcessingLoadAddressPrivateKey(ctx, item.Address, item.WalletAddressesID, dto.IP)
		if err != nil {
			s.logger.Error("Error store DB log for load private key", err)
		}
	}

	return bb, nil
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
	return nil
}

func (s *Service) calculateAddressHash(addresses []*models.WalletAddress) string {
	hashes := lo.Map(addresses, func(w *models.WalletAddress, _ int) string {
		h := sha256.New()
		if _, err := h.Write([]byte(w.Address)); err != nil {
			s.logger.Error("failed to hash address", err)
			return ""
		}
		return hex.EncodeToString(h.Sum(nil))[:6]
	})
	return strings.Join(lo.Uniq(hashes), "-")
}

func (s *Service) generateWalletAddresses(ctx context.Context, tx pgx.Tx, owner *models.User, wallet *models.Wallet, str *models.Store, currencies []*models.Currency, amount string) ([]*models.WalletAddress, error) {
	result := make([]*models.WalletAddress, 0, len(currencies))

	for _, c := range currencies {
		if c.IsFiat {
			continue
		}

		addr, err := s.getOrCreateWalletAddress(ctx, tx, owner, wallet, c)
		if err != nil {
			return nil, fmt.Errorf("failed to get or create wallet address: %w", err)
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

func sortWallets(wallets []*ProcessingWalletWithAssets) {
	sort.Slice(wallets, func(i, j int) bool {
		pi, iOk := models.BlockchainSortOrder[wallets[i].Blockchain]
		pj, jOk := models.BlockchainSortOrder[wallets[j].Blockchain]

		if iOk && jOk {
			return pi < pj
		}
		if iOk {
			return true
		}
		if jOk {
			return false
		}
		return string(wallets[i].Blockchain) < string(wallets[j].Blockchain)
	})
}
