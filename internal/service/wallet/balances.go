package wallet

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/dv-net/dv-merchant/internal/delivery/http/request/wallet_request"
	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/internal/service/processing"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_wallet_addresses"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/internal/util"
	addressesv2 "github.com/dv-net/dv-proto/gen/go/eproxy/addresses/v2"
	evmv2 "github.com/dv-net/dv-proto/gen/go/eproxy/evm/v2"
	"github.com/jackc/pgx/v5"
	"github.com/puzpuzpuz/xsync/v3"
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
			Dirty:           w.Dirty,
		})
	}

	return &storecmn.FindResponseWithFullPagination[*models.WalletWithUSDBalance]{
		Items:      addresses,
		Pagination: wallets.Pagination,
	}, nil
}

func (s *Service) GetHotWalletsTotalBalance(ctx context.Context, user *models.User) (*AddressesTotalBalance, error) {
	rates, err := s.exrateService.LoadRatesList(ctx, user.RateSource.String())
	if err != nil {
		return nil, fmt.Errorf("failed to load rates: %w", err)
	}

	result, err := s.storage.WalletAddresses().GetHotWalletsTotalBalanceWithDust(ctx, repo_wallet_addresses.GetHotWalletsTotalBalanceWithDustParams{
		UserID:       user.ID,
		CurrencyIds:  rates.CurrencyIDs,
		CurrencyRate: rates.Rate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get hot wallets balance: %w", err)
	}

	return &AddressesTotalBalance{
		TotalUSD:  result.TotalUsd,
		TotalDust: result.DustUsd,
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
				s.logger.Errorw("failed to process blockchain wallets", "error", err, "blockchain", blockchain.String())
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

const (
	TRC20EnergyPriceTRX    = 28
	TRC20BandwidthPriceTRX = 1
	TRXBandwidthPriceTRX   = 0.348
)

// ChainConfig holds chain-specific gas parameters
type ChainConfig struct {
	NativeTransferGas   int64
	ERC20TransferGas    int64
	IsL2                bool
	L1DataFeeMultiplier float64
}

var ChainConfigs = map[models.Blockchain]ChainConfig{
	models.BlockchainEthereum:          {NativeTransferGas: 21000, ERC20TransferGas: 65000},
	models.BlockchainArbitrum:          {NativeTransferGas: 21000, ERC20TransferGas: 80000, IsL2: true, L1DataFeeMultiplier: 1.5},
	models.BlockchainBinanceSmartChain: {NativeTransferGas: 21000, ERC20TransferGas: 60000},
	models.BlockchainPolygon:           {NativeTransferGas: 21000, ERC20TransferGas: 65000},
}

func (s *Service) processBlockchainWallets(ctx context.Context, blockchain models.Blockchain, dto GetProcessingWalletsDTO, enabledCurrencies []*models.Currency) ([]*ProcessingWalletWithAssets, error) {
	wallets, err := s.processingService.GetOwnerProcessingWallets(ctx, processing.GetOwnerProcessingWalletsParams{
		OwnerID:    dto.OwnerID,
		Blockchain: util.Pointer(blockchain),
	})
	if err != nil {
		if strings.Contains(err.Error(), "not available") {
			return nil, nil
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

	cur, exists := lo.Find(enabledCurrencies, func(c *models.Currency) bool {
		return c.ID == nativeCurrency && c.Blockchain != nil && *c.Blockchain == blockchain
	})
	if !exists {
		return nil, fmt.Errorf("native currency %s for blockchain %s is not enabled", nativeCurrency, blockchain)
	}

	processingWallet := &ProcessingWalletWithAssets{
		Address:    wallet.Address,
		Blockchain: wallet.Blockchain,
		Currency: &models.CurrencyShort{
			ID:            cur.ID,
			Code:          cur.Code,
			Precision:     cur.Precision,
			Name:          cur.Name,
			Blockchain:    cur.Blockchain,
			IsEVMLike:     cur.Blockchain.IsEVMLike(),
			IsBitcoinLike: cur.Blockchain.IsBitcoinLike(),
			IsStableCoin:  cur.IsStablecoin,
		},
	}

	if err := s.assembleWalletBalance(ctx, processingWallet, wallet, blockchain, currencies, enabledCurrencies); err != nil {
		return nil, fmt.Errorf("failed to assemble wallet balance: %w", err)
	}

	if lo.Contains(currencies, nativeCurrency) {
		if err := s.addBlockchainSpecificData(ctx, processingWallet, wallet, blockchain); err != nil {
			return nil, fmt.Errorf("failed to add blockchain specific data: %w", err)
		}
	}
	return processingWallet, nil
}

func (s *Service) assembleWalletBalance(ctx context.Context, processingWallet *ProcessingWalletWithAssets, wallet processing.WalletProcessing, blockchain models.Blockchain, currencies []string, enabledCurrencies []*models.Currency) error {
	nativeCurrency, err := blockchain.NativeCurrency()
	if err != nil {
		return fmt.Errorf("failed to get native currency for blockchain %s: %w", blockchain, err)
	}

	assets, err := s.assembleAssets(ctx, blockchain, currencies, enabledCurrencies, wallet.Assets)
	if err != nil {
		return fmt.Errorf("failed to assemble assets: %w", err)
	}
	processingWallet.Assets = assets

	asset, exists := lo.Find(assets, func(a *Asset) bool { return nativeCurrency == a.CurrencyID })
	if exists {
		processingWallet.Balance = &Balance{NativeToken: asset.Amount, NativeTokenUSD: asset.AmountUSD}
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

	tronData.MaxTransfersNative = tokenBalance.Div(decimal.NewFromFloat(TRXBandwidthPriceTRX)).Floor().String()
	tronData.MaxTransfersTRC20 = tokenBalance.Div(decimal.NewFromFloat(TRC20EnergyPriceTRX).Add(decimal.NewFromFloat(TRC20BandwidthPriceTRX))).Floor().String()
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

	evmData := &EVMData{SuggestedGasPrice: estimatedGasPrice.Msg.GasFeeWei}

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

	nativeTransferCost := gasPriceEth.Mul(decimal.NewFromInt(chainConfig.NativeTransferGas))
	erc20TransferCost := gasPriceEth.Mul(decimal.NewFromInt(chainConfig.ERC20TransferGas))

	if chainConfig.IsL2 {
		l1DataFeeNative := nativeTransferCost.Mul(decimal.NewFromFloat(chainConfig.L1DataFeeMultiplier))
		l1DataFeeERC20 := erc20TransferCost.Mul(decimal.NewFromFloat(chainConfig.L1DataFeeMultiplier))
		nativeTransferCost = nativeTransferCost.Add(l1DataFeeNative)
		erc20TransferCost = erc20TransferCost.Add(l1DataFeeERC20)
		evmData.L1DataFeeEstimate = l1DataFeeNative.String()
	}

	var maxNativeTransfers, maxERC20Transfers int64
	if !nativeTransferCost.IsZero() {
		maxNativeTransfers = balance.Div(nativeTransferCost).IntPart()
	}
	if !erc20TransferCost.IsZero() {
		maxERC20Transfers = balance.Div(erc20TransferCost).IntPart()
	}

	evmData.IsL2 = chainConfig.IsL2
	evmData.MaxTransfersNative = fmt.Sprintf("%d", maxNativeTransfers)
	evmData.MaxTransfersERC20 = fmt.Sprintf("%d", maxERC20Transfers)
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
		for _, cur := range currencies {
			if c == cur.ID {
				return c, true
			}
		}
		return "", false
	})

	assets := make([]*Asset, 0, len(filterCurrencies))
	for _, cur := range enabledCurrencies {
		if !slices.Contains(filterCurrencies, cur.ID) {
			continue
		}

		asset, exists := lo.Find(processingAssets, func(a *processing.Asset) bool {
			return a.Identity == cur.ID || a.Identity == cur.ContractAddress.String
		})
		if exists {
			amountUSD, err := s.currConvService.Convert(ctx, currconv.ConvertDTO{
				Source:     models.ExchangeSlugBinance.String(),
				From:       cur.Code,
				To:         models.CurrencyCodeUSDT,
				Amount:     asset.Amount,
				StableCoin: false,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to convert %s %s to USDT: %w", asset.Amount, asset.Identity, err)
			}
			assets = append(assets, &Asset{CurrencyID: cur.ID, Identity: cur.Code, Amount: asset.Amount, AmountUSD: amountUSD.String()})
		} else {
			assets = append(assets, &Asset{CurrencyID: cur.ID, Amount: "0", AmountUSD: "0", Identity: cur.Code})
		}
	}
	return assets, nil
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
		s.logger.Errorw("fetch wallet balance", "error", err)
	}

	totalUSD := decimal.Zero
	total.Range(func(_ string, v decimal.Decimal) bool {
		totalUSD = totalUSD.Add(v)
		return true
	})
	return &AddressesTotalBalance{TotalUSD: totalUSD}, nil
}

func (s *Service) processWalletBalance(ctx context.Context, user *models.User, wallet CalcBalanceDTO) (decimal.Decimal, error) {
	blockchain, err := wallet.Blockchain.ToEPb()
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to convert blockchain [%s] to pb: %w", blockchain, err)
	}

	walletsRes, err := s.eproxyService.Wallets().Info(ctx, connect.NewRequest(&addressesv2.InfoRequest{
		Address:    wallet.Address,
		Blockchain: blockchain,
	}))
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

func (s *Service) filterProxyEnabledAssets(_ context.Context, currencies []*models.Currency, assets []*addressesv2.AssetInfo) []*addressesv2.AssetInfo {
	res := make([]*addressesv2.AssetInfo, 0)
	for _, asset := range assets {
		if slices.IndexFunc(currencies, func(c *models.Currency) bool {
			return strings.EqualFold(asset.AssetIdentifier, c.ContractAddress.String)
		}) != -1 {
			res = append(res, asset)
		}
	}
	return res
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
