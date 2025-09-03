package binance

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_chains"
	"github.com/dv-net/dv-merchant/internal/tools/hash"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/binance"
	binancemodels "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/models"
	binancerequests "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrSymbolTradingHalted = errors.New("symbol trading is halted")
)

func NewService(apiKey, secretKey string, public bool, baseURL *url.URL, storage storage.IStorage, convSvc currconv.ICurrencyConvertor) (*Service, error) {
	exClient, err := binance.NewBaseClient(&binance.ClientOptions{
		APIKey:       apiKey,
		SecretKey:    secretKey,
		BaseURL:      baseURL,
		PublicClient: public,
	})
	if err != nil {
		return nil, err
	}

	connHash, err := hash.SHA256ConnectionHash(models.ExchangeSlugBinance.String(), apiKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection hash: %w", err)
	}

	return &Service{
		exClient: exClient,
		storage:  storage,
		convSvc:  convSvc,
		connHash: connHash,
	}, nil
}

type Service struct {
	exClient *binance.BaseClient
	storage  storage.IStorage
	convSvc  currconv.ICurrencyConvertor
	connHash string
}

func (o *Service) TestConnection(ctx context.Context) error {
	if _, err := o.exClient.Spot().AccountInformation(ctx, &binancerequests.AccountInformationRequest{}); err != nil {
		return err
	}
	return nil
}

type UniversalBalanceDTO struct {
	Balance decimal.Decimal `json:"balance"`
	Ccy     string          `json:"ccy"`
}

func (o *Service) GetAccountBalance(ctx context.Context) ([]*models.AccountBalanceDTO, error) {
	assetBalances, err := o.exClient.Wallet().GetSpotAssets(ctx, &binancerequests.GetSpotAssetsRequest{})
	if err != nil {
		return nil, err
	}
	fundingBalances, err := o.exClient.Wallet().GetFundingAssets(ctx, &binancerequests.GetFundingAssetsRequest{})
	if err != nil {
		return nil, err
	}

	balanceMap := make(map[string]*UniversalBalanceDTO)
	for _, balance := range assetBalances.Data {
		amount, err := decimal.NewFromString(balance.Free)
		if err != nil {
			return nil, err
		}
		if record, exists := balanceMap[balance.Asset]; exists {
			record.Balance = record.Balance.Add(amount)
		} else {
			balanceMap[balance.Asset] = &UniversalBalanceDTO{
				Ccy:     balance.Asset,
				Balance: amount,
			}
		}
	}

	for _, balance := range fundingBalances.Data {
		amount, err := decimal.NewFromString(balance.Free)
		if err != nil {
			return nil, err
		}
		if record, exists := balanceMap[balance.Asset]; exists {
			record.Balance = record.Balance.Add(amount)
		} else {
			balanceMap[balance.Asset] = &UniversalBalanceDTO{
				Ccy:     balance.Asset,
				Balance: amount,
			}
		}
	}

	balances := make([]*models.AccountBalanceDTO, 0, len(balanceMap))
	for _, balance := range balanceMap {
		currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByTicker(ctx, balance.Ccy)
		if err != nil {
			continue
		}
		amountUSD, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
			Source:     "binance",
			From:       balance.Ccy,
			To:         "USDT",
			Amount:     balance.Balance.String(),
			StableCoin: false,
		})
		if err != nil {
			return nil, err
		}
		balances = append(balances, &models.AccountBalanceDTO{
			Currency:  currencyID,
			Amount:    balance.Balance,
			AmountUSD: amountUSD.Round(4),
			Type:      models.CurrencyTypeCrypto.String(),
		})
	}

	return balances, nil
}

func (o *Service) GetCurrencyBalance(ctx context.Context, currency string) (*decimal.Decimal, error) {
	spotBalances, err := o.exClient.Wallet().GetSpotAssets(ctx, &binancerequests.GetSpotAssetsRequest{
		Asset: currency,
	})
	if err != nil {
		return nil, err
	}

	fundingBalances, err := o.exClient.Wallet().GetFundingAssets(ctx, &binancerequests.GetFundingAssetsRequest{
		Asset: currency,
	})
	if err != nil {
		return nil, err
	}

	spotAmt, fundingAmt := decimal.Zero, decimal.Zero
	if len(fundingBalances.Data) > 0 {
		fundingAmt, err = decimal.NewFromString(fundingBalances.Data[0].Free)
		if err != nil {
			return nil, err
		}
	}

	if len(spotBalances.Data) > 0 {
		spotAmt, err = decimal.NewFromString(spotBalances.Data[0].Free)
		if err != nil {
			return nil, err
		}
	}
	amt := spotAmt.Add(fundingAmt)
	return &amt, nil
}

func (o *Service) GetExchangeSymbols(ctx context.Context) ([]*models.ExchangeSymbolDTO, error) {
	res, err := o.exClient.MarketData().GetExchangeInfo(ctx, &binancerequests.GetExchangeInfoRequest{})
	if err != nil {
		return nil, err
	}

	dto := make([]*models.ExchangeSymbolDTO, 0, len(res.Symbols)*2)
	for _, s := range res.Symbols {
		if s.Status != binancemodels.SymbolStatusTrading {
			continue
		}
		base, quote := strings.ToUpper(s.BaseAsset), strings.ToUpper(s.QuoteAsset)
		dto = append(dto, &models.ExchangeSymbolDTO{
			Symbol:      s.Symbol,
			DisplayName: base + "/" + quote,
			BaseSymbol:  s.BaseAsset,
			QuoteSymbol: s.QuoteAsset,
			Type:        "sell",
		}, &models.ExchangeSymbolDTO{
			Symbol:      s.Symbol,
			DisplayName: quote + "/" + base,
			BaseSymbol:  s.BaseAsset,
			QuoteSymbol: s.QuoteAsset,
			Type:        "buy",
		})
	}

	return dto, nil
}

func (o *Service) GetDepositAddresses(ctx context.Context, currency, chain string) ([]*models.DepositAddressDTO, error) {
	enabledCurrencies, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugBinance)
	if err != nil {
		return nil, fmt.Errorf("fetch enabled currencies: %w", err)
	}
	network, found := lo.Find(enabledCurrencies, func(i *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
		return i.Ticker == currency && i.Chain == chain
	})
	if !found {
		return nil, fmt.Errorf("currency %s network cannot be found", currency)
	}

	exchangeAddress, err := o.exClient.Wallet().GetDefaultDepositAddress(ctx, &binancerequests.GetDefaultDepositAddressRequest{
		Coin:    currency,
		Network: network.Chain,
	})
	if err != nil {
		return nil, fmt.Errorf("get deposit addresses: %w", err)
	}
	currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByParams(ctx, repo_exchange_chains.GetCurrencyIDByParamsParams{
		Ticker: network.Ticker,
		Chain:  network.Chain,
		Slug:   models.ExchangeSlugBinance,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get internal currency id %s: %w", network.Chain, err)
	}
	return []*models.DepositAddressDTO{
		{
			Address:          exchangeAddress.Address,
			Currency:         currencyID,
			Chain:            network.Chain,
			AddressType:      models.DepositAddress,
			InternalCurrency: exchangeAddress.Coin,
		},
	}, nil
}

func (o *Service) CreateWithdrawalOrder(ctx context.Context, args *models.CreateWithdrawalOrderParams) (*models.ExchangeWithdrawalDTO, error) {
	args.NativeAmount = args.NativeAmount.RoundDown(int32(args.WithdrawalPrecision)) //nolint:gosec

	internalCurrencyID, err := o.storage.ExchangeChains().GetTickerByCurrencyID(ctx, repo_exchange_chains.GetTickerByCurrencyIDParams{
		CurrencyID: args.Currency,
		Slug:       models.ExchangeSlugBinance,
	})
	if err != nil {
		return nil, err
	}

	funding, err := o.exClient.Wallet().GetFundingAssets(ctx, &binancerequests.GetFundingAssetsRequest{
		Asset: internalCurrencyID,
	})
	if err != nil {
		return nil, err
	}

	spot, err := o.exClient.Wallet().GetSpotAssets(ctx, &binancerequests.GetSpotAssetsRequest{
		Asset: internalCurrencyID,
	})
	if err != nil {
		return nil, err
	}

	spotBalance := o.getSpotBalance(internalCurrencyID, spot.Data).RoundDown(int32(args.WithdrawalPrecision))          //nolint:gosec
	fundingBalance := o.getFundingBalance(internalCurrencyID, funding.Data).RoundDown(int32(args.WithdrawalPrecision)) //nolint:gosec

	totalBalance := spotBalance.Add(fundingBalance)

	if spotBalance.LessThan(args.NativeAmount) {
		if spotBalance.Add(fundingBalance).LessThan(args.NativeAmount) {
			return nil, ErrInsufficientBalance
		}

		transferAmount := totalBalance.Sub(spotBalance)
		transferAmount = transferAmount.RoundDown(int32(args.WithdrawalPrecision)) //nolint:gosec
		_, err = o.exClient.Wallet().UniversalTransfer(ctx, &binancerequests.UniversalTransferRequest{
			Asset:  internalCurrencyID,
			Type:   binancemodels.TransferTypeFundingToSpot,
			Amount: transferAmount.String(),
		})
		if err != nil {
			return nil, err
		}
	}

	req := &binancerequests.WithdrawalRequest{
		Coin:    internalCurrencyID,
		Network: args.Chain,
		Address: args.Address,
		Amount:  args.NativeAmount.String(),
	}

	clientOrderID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	req.WithdrawOrderId = strings.ReplaceAll(clientOrderID.String(), "-", "")

	order, err := o.exClient.Wallet().Withdrawal(ctx, req)
	if err != nil {
		return nil, err
	}

	dto := &models.ExchangeWithdrawalDTO{
		InternalOrderID: req.WithdrawOrderId,
		ExternalOrderID: order.ID,
	}

	return dto, nil
}

func (o *Service) CreateSpotOrder(ctx context.Context, from string, to string, side string, ticker string, amount *decimal.Decimal, rule *models.OrderRulesDTO) (*models.ExchangeOrderDTO, error) {
	side = strings.ToUpper(side)

	orderID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	clientOrderID := strings.ReplaceAll(orderID.String(), "-", "")

	symbolsData, err := o.exClient.MarketData().GetExchangeInfo(ctx, &binancerequests.GetExchangeInfoRequest{
		Symbol: ticker,
	})
	if err != nil {
		return nil, err
	}
	if len(symbolsData.Symbols) == 0 {
		return nil, fmt.Errorf("symbol %s not found", ticker)
	}
	symbolData := symbolsData.Symbols[0]
	if symbolData.Status != binancemodels.SymbolStatusTrading {
		return nil, fmt.Errorf("symbol %s cannot be traded with status %s", ticker, symbolData.Status.String())
	}

	tSpotOrderRequest := &binancerequests.TestNewOrderRequest{
		NewOrderRequest: binancerequests.NewOrderRequest{
			Symbol:           ticker,
			Side:             side,
			Type:             binancemodels.OrderTypeMarket.String(),
			NewClientOrderId: clientOrderID,
		},
		ComputeCommissionRates: true,
	}

	switch side {
	case binancemodels.OrderSideBuy.String():
		from = symbolData.QuoteAsset
		to = symbolData.BaseAsset
	case binancemodels.OrderSideSell.String():
		from = symbolData.BaseAsset
		to = symbolData.QuoteAsset
	}

	orderMinimumQuote := rule.MinOrderValue
	orderMinimumBase := rule.MinOrderAmount

	// orderMinimumQuote := rule.NotionalFilter.MinNotional
	// orderMinimumBase := rule.LotSizeFilter.MinQty

	universalTransferRequest := &binancerequests.UniversalTransferRequest{
		Type: binancemodels.TransferTypeFundingToSpot,
	}

	spotBalances, err := o.exClient.Wallet().GetSpotAssets(ctx, &binancerequests.GetSpotAssetsRequest{
		Asset: from,
	})
	if err != nil {
		return nil, err
	}

	fundingBalances, err := o.exClient.Wallet().GetFundingAssets(ctx, &binancerequests.GetFundingAssetsRequest{
		Asset: from,
	})
	if err != nil {
		return nil, err
	}

	maxAmount := decimal.Zero  //nolint:staticcheck
	spotAmt := decimal.Zero    //nolint:staticcheck
	fundingAmt := decimal.Zero //nolint:staticcheck

	universalTransferRequest.Asset = from
	spotAmt = o.getSpotBalance(from, spotBalances.Data)
	fundingAmt = o.getFundingBalance(from, fundingBalances.Data)
	maxAmount = fundingAmt.Add(spotAmt)

	switch tSpotOrderRequest.Side {
	case binancemodels.OrderSideBuy.String():
		orderMinQuote, err := decimal.NewFromString(orderMinimumQuote)
		if err != nil {
			return nil, err
		}
		if maxAmount.LessThan(orderMinQuote) {
			return nil, ErrInsufficientBalance
		}
	case binancemodels.OrderSideSell.String():
		orderMinBase, err := decimal.NewFromString(orderMinimumBase)
		if err != nil {
			return nil, err
		}
		if maxAmount.LessThan(orderMinBase) {
			return nil, ErrInsufficientBalance
		}
	default:
		return nil, fmt.Errorf("unsupported order type %s", tSpotOrderRequest.Side)
	}

	remainingToTransfer := maxAmount.Sub(spotAmt)
	if remainingToTransfer.GreaterThan(decimal.Zero) {
		remainder := remainingToTransfer.String()
		universalTransferRequest.Amount = remainder

		_, err = o.exClient.Wallet().UniversalTransfer(ctx, universalTransferRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to transfer funds: %w", err)
		}
	}

	filters, err := utils.ExtractMarketFilters(symbolData.Filters)
	if err != nil {
		return nil, err
	}

	precision := utils.ConvertPrecision(filters.LotSizeFilter.MinQty.String())

	if from == symbolData.BaseAsset {
		tSpotOrderRequest.NewOrderRequest.Quantity = maxAmount.RoundDown(int32(precision.IntPart())).String()
	} else {
		tSpotOrderRequest.NewOrderRequest.QuoteOrderQty = maxAmount.RoundDown(int32(precision.IntPart())).String()
	}

	order, err := o.exClient.Spot().NewOrder(ctx, &tSpotOrderRequest.NewOrderRequest)
	if err != nil {
		return nil, err
	}

	return &models.ExchangeOrderDTO{
		ExchangeOrderID: strconv.Itoa(order.OrderId),
		ClientOrderID:   order.ClientOrderId,
		Amount:          order.ExecutedQty,
	}, nil
}

func (o *Service) getSpotBalance(symbol string, spot []binancemodels.AssetBalance) decimal.Decimal {
	spotAmount := decimal.Zero
	spotBalance, exists := lo.Find(spot, func(item binancemodels.AssetBalance) bool {
		return strings.EqualFold(item.Asset, symbol)
	})
	if exists {
		amount, err := decimal.NewFromString(spotBalance.Free)
		if err != nil {
			return decimal.Zero
		}
		spotAmount = spotAmount.Add(amount)
	}
	return spotAmount
}

func (o *Service) getFundingBalance(symbol string, funding []binancemodels.AssetBalance) decimal.Decimal {
	fundingAmount := decimal.Zero
	fundingBalance, exists := lo.Find(funding, func(item binancemodels.AssetBalance) bool {
		return strings.EqualFold(item.Asset, symbol)
	})
	if exists {
		amount, err := decimal.NewFromString(fundingBalance.Free)
		if err != nil {
			return decimal.Zero
		}
		fundingAmount = fundingAmount.Add(amount)
	}
	return fundingAmount
}

func (o *Service) GetOrderRule(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	res, err := o.exClient.MarketData().GetExchangeInfo(ctx, &binancerequests.GetExchangeInfoRequest{
		Symbol: ticker,
	})
	if err != nil {
		return nil, err
	}
	if len(res.Symbols) == 0 {
		return nil, errors.New("symbol not found")
	}
	symbolData := res.Symbols[0]

	if symbolData.Status == binancemodels.SymbolStatusBreak || symbolData.Status == binancemodels.SymbolStatusHalt {
		return nil, fmt.Errorf("symbol %s is not available for trading: %w", ticker, ErrSymbolTradingHalted)
	}

	filters, err := utils.ExtractMarketFilters(symbolData.Filters)
	if err != nil {
		return nil, err
	}

	minOrderAmount := decimal.Zero
	minOrderValue := filters.NotionalFilter.MinNotional

	precision := utils.ConvertPrecision(filters.LotSizeFilter.MinQty.String())

	// FIXME: As of right now, MIN_NOTIONAL reflects min volume of a single trade on symbol.
	// MinOrderAmount should reflect minimum base asset amount calculated from MinNotional converted to base asset.
	// MinOrderValue should reflect minimum quote asset amount calculated from MinNotional converted to quote asset.
	// Also add checks agains filters to not only apply to the volume rules, but min order qty filter as well.
	basePriceData, err := o.exClient.MarketData().GetSymbolPriceTicker(ctx, &binancerequests.GetSymbolPriceTickerRequest{
		Symbol: symbolData.BaseAsset + "USDT",
	})
	if err != nil {
		return nil, err
	}
	basePrice, err := decimal.NewFromString(basePriceData.Price)
	if err != nil {
		return nil, err
	}

	minOrderAmount = filters.NotionalFilter.MinNotional.Div(basePrice).RoundUp(int32(precision.IntPart()))
	if symbolData.QuoteAsset != "USDT" {
		quotePriceData, err := o.exClient.MarketData().GetSymbolPriceTicker(ctx, &binancerequests.GetSymbolPriceTickerRequest{
			Symbol: symbolData.QuoteAsset + "USDT",
		})
		if err != nil {
			return nil, err
		}
		quotePrice, err := decimal.NewFromString(quotePriceData.Price)
		if err != nil {
			return nil, err
		}
		minOrderValue = filters.NotionalFilter.MinNotional.Div(quotePrice).RoundUp(int32(precision.IntPart()))
	}

	dto := &models.OrderRulesDTO{
		Symbol:          symbolData.Symbol,
		State:           symbolData.Status.String(),
		BaseCurrency:    symbolData.BaseAsset,
		QuoteCurrency:   symbolData.QuoteAsset,
		MinOrderAmount:  minOrderAmount.String(),
		MaxOrderAmount:  filters.LotSizeFilter.MaxQty.String(),
		MinOrderValue:   minOrderValue.String(),
		AmountPrecision: symbolData.BaseAssetPrecision,
		ValuePrecision:  symbolData.QuoteAssetPrecision,
	}
	return dto, nil
}

func (o *Service) GetOrderDetails(ctx context.Context, args *models.GetOrderByIDParams) (*models.OrderDetailsDTO, error) {
	order := &models.OrderDetailsDTO{
		State:     models.ExchangeOrderStatusFailed,
		Amount:    decimal.Zero,
		AmountUSD: decimal.Zero,
	}
	req := binancerequests.QueryOrderRequest{}

	if args.InstrumentID != nil {
		req.Symbol = *args.InstrumentID
	}

	if args.ExternalOrderID != nil {
		orderID, err := strconv.Atoi(*args.ExternalOrderID)
		if err != nil {
			return nil, err
		}
		req.OrderID = int64(orderID) //nolint:gosec
	}
	if args.ClientOrderID != nil {
		req.OrigClientOrderId = *args.ClientOrderID
	}

	res, err := o.exClient.Spot().QueryOrder(ctx, &req)
	if err != nil {
		return nil, err
	}

	amt, err := decimal.NewFromString(res.ExecutedQty)
	if err != nil {
		return nil, err
	}

	order.Amount = amt

	symbolInfo, err := o.exClient.MarketData().GetExchangeInfo(ctx, &binancerequests.GetExchangeInfoRequest{
		Symbol: res.Symbol,
	})
	if err != nil {
		return nil, err
	}

	if len(symbolInfo.Symbols) == 0 {
		return nil, fmt.Errorf("symbol %s not found", res.Symbol)
	}
	symbol := symbolInfo.Symbols[0]

	// Calculate for */USDT symbol
	if strings.EqualFold(symbol.QuoteAsset, "USDT") {
		priceData, err := o.exClient.MarketData().GetSymbolPriceTicker(ctx, &binancerequests.GetSymbolPriceTickerRequest{
			Symbol: res.Symbol,
		})
		if err != nil {
			return nil, err
		}

		price, err := decimal.NewFromString(priceData.Price)
		if err != nil {
			return nil, err
		}

		order.AmountUSD = price.Mul(order.Amount)
	}

	// Calculate for */* symbol, non-USDT
	{
		basePriceData, err := o.exClient.MarketData().GetSymbolPriceTicker(ctx, &binancerequests.GetSymbolPriceTickerRequest{
			Symbol: symbol.BaseAsset + "USDT",
		})
		if err != nil {
			return nil, err
		}
		price, err := decimal.NewFromString(basePriceData.Price)
		if err != nil {
			return nil, err
		}

		order.AmountUSD = price.Mul(order.Amount)
	}

	switch res.Status {
	case binancemodels.OrderStatusFilled.String():
		order.State = models.ExchangeOrderStatusCompleted
	case binancemodels.OrderStatusCanceled.String(), binancemodels.OrderStatusRejected.String(), binancemodels.OrderStatusExpired.String():
		order.State = models.ExchangeOrderStatusFailed
	default:
		order.State = models.ExchangeOrderStatusInProgress
	}

	return order, nil
}

func (o *Service) GetWithdrawalRules(ctx context.Context, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	currEnabled, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugBinance)
	if err != nil {
		return nil, err
	}

	currEnabled = lo.Filter(currEnabled, func(i *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) bool {
		return lo.Contains(currencies, i.ID.String)
	})

	coinInformation, err := o.exClient.Wallet().GetAllCoinInformation(ctx)
	if err != nil {
		return nil, err
	}

	enabledNetworks := lo.FlatMap(coinInformation.Data, func(i *binancemodels.CoinInfo, _ int) []binancemodels.Network {
		return lo.Filter(i.NetworkList, func(n binancemodels.Network, _ int) bool {
			return lo.ContainsBy(currEnabled, func(item *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
				return item.Chain == n.Network && item.Ticker == n.Coin
			})
		})
	})

	result := lo.Map(enabledNetworks, func(n binancemodels.Network, _ int) *models.WithdrawalRulesDTO {
		dto := &models.WithdrawalRulesDTO{
			Currency:            n.Coin,
			Chain:               n.Network,
			MinDepositAmount:    n.DepositDust,
			MinWithdrawAmount:   n.WithdrawMin,
			MaxWithdrawAmount:   n.WithdrawMax,
			NumOfConfirmations:  strconv.Itoa(n.MinConfirm),
			WithdrawFeeType:     models.WithdrawalFeeTypeFixed,
			WithdrawPrecision:   utils.ConvertPrecision(n.WithdrawIntegerMultiple).String(),
			WithdrawQuotaPerDay: "",
			Fee:                 n.WithdrawFee,
		}
		return dto
	})
	return result, nil
}

func (o *Service) GetWithdrawalByID(ctx context.Context, args *models.GetWithdrawalByIDParams) (*models.WithdrawalStatusDTO, error) {
	request := &binancerequests.GetWithdrawalHistoryRequest{}
	if args.ClientOrderID != nil {
		request.WithdrawOrderId = *args.ClientOrderID
	}
	res, err := o.exClient.Wallet().GetWithdrawalHistory(ctx, request)
	if err != nil {
		return nil, err
	}
	if len(res.Data) == 0 {
		return nil, errors.New("withdrawal not found")
	}
	withdrawal := res.Data[0]
	dto := &models.WithdrawalStatusDTO{
		ID:     withdrawal.WithdrawOrderId,
		Status: withdrawal.Status.String(),
	}
	if withdrawal.TxId != "" {
		dto.TxHash = withdrawal.TxId
	}
	return dto, nil
}

func (o *Service) GetOrderRules(ctx context.Context, tickers ...string) ([]*models.OrderRulesDTO, error) {
	rules := make([]*models.OrderRulesDTO, 0, len(tickers))
	for _, ticker := range tickers {
		rule, err := o.GetOrderRule(ctx, ticker)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (o *Service) GetConnectionHash() string {
	return o.connHash
}
