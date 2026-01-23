package bybit

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_chains"
	"github.com/dv-net/dv-merchant/internal/tools/hash"
	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bybit"
	bybitmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/models"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/dv-merchant/pkg/retry"
)

var (
	ErrInsufficientBalance         = errors.New("insufficient balance")
	ErrUnprocessableCurrencyStatus = errors.New("unprocessable currency status")
	ErrMaxOrderValueReached        = errors.New("max order value reached")
)

const (
	WithdrawalStep = 10
)

type Service struct {
	exClient *bybit.BaseClient
	storage  storage.IStorage
	convSvc  currconv.ICurrencyConvertor
	l        logger.Logger
	connHash string
}

func NewService(logger logger.Logger, accessKey, secretKey string, baseURL *url.URL, storage storage.IStorage, store limiter.Store, convSvc currconv.ICurrencyConvertor) (*Service, error) {
	exClient := bybit.NewBaseClient(&bybit.ClientOptions{
		KeyAPI:    accessKey,
		KeySecret: secretKey,
		BaseURL:   baseURL,
	}, store, bybit.WithLogger(logger))

	connHash, err := hash.SHA256ConnectionHash(models.ExchangeSlugBybit.String(), accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection hash: %w", err)
	}

	return &Service{
		l:        logger,
		exClient: exClient,
		storage:  storage,
		convSvc:  convSvc,
		connHash: connHash,
	}, nil
}

func (o *Service) TestConnection(ctx context.Context) error {
	resp, err := o.exClient.Account().GetAccountInfo(ctx)
	if err != nil {
		return err
	}
	if resp.Result.UnifiedMarginStatus != bybitmodels.UnifiedMarginStatusUnifiedTradingAccount20 {
		retrier := retry.New(retry.WithContext(ctx), retry.WithDelay(time.Second))
		return retrier.Do(func() error {
			resp, err := o.exClient.Account().UpgradeToUnifiedTradingAccount(ctx)
			if err != nil {
				return fmt.Errorf("failed to upgrade to unified trading account: %w", err)
			}
			if resp.Result.UnifiedUpdateStatus == bybitmodels.UnifiedUpgradeStatusSuccess {
				return nil
			}
			if resp.Result.UnifiedUpdateStatus == bybitmodels.UnifiedUpgradeStatusFail {
				return fmt.Errorf("failed to upgrade to unified trading account: %s", resp.RetMsg)
			}
			if resp.Result.UnifiedUpdateStatus == bybitmodels.UnifiedUpgradeStatusProcess {
				return fmt.Errorf("upgrade to unified trading account is still processing")
			}
			return nil
		})
	}
	return nil
}

func (o *Service) GetAccountBalance(ctx context.Context) ([]*models.AccountBalanceDTO, error) {
	enabledCurrencies, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugBybit)
	if err != nil {
		return nil, err
	}

	tradingBalances, err := o.exClient.Account().GetTradingBalance(ctx, &requests.GetTradingBalanceRequest{
		AccountType: bybitmodels.AccountTypeUnified.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get trading balance: %w", err)
	}

	fundingBalances, err := o.exClient.Account().GetAllCoinsBalance(ctx, &requests.GetFundingBalanceRequest{
		AccountType: bybitmodels.AccountTypeFund.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get funding balance: %w", err)
	}

	tradingBalanceMap := make(map[string]decimal.Decimal)
	if len(tradingBalances.Result.List) > 0 && len(tradingBalances.Result.List[0].Coin) > 0 {
		for _, coin := range tradingBalances.Result.List[0].Coin {
			if coin.WalletBalance == "" {
				tradingBalanceMap[coin.Coin] = decimal.Zero
			} else {
				balance, _ := decimal.NewFromString(coin.WalletBalance)
				tradingBalanceMap[coin.Coin] = balance
			}
		}
	}

	fundingBalanceMap := make(map[string]decimal.Decimal)
	if len(fundingBalances.Result.Balance) > 0 {
		for _, balance := range fundingBalances.Result.Balance {
			if balance.WalletBalance == "" {
				fundingBalanceMap[balance.Coin] = decimal.Zero
			} else {
				bal, _ := decimal.NewFromString(balance.WalletBalance)
				fundingBalanceMap[balance.Coin] = bal
			}
		}
	}

	// Deduplicate by ticker to avoid counting the same balance multiple times across different chains
	seenTickers := make(map[string]bool)
	balances := make([]*models.AccountBalanceDTO, 0)

	for _, currency := range enabledCurrencies {
		// Skip if we've already processed this ticker
		if seenTickers[currency.Ticker] {
			continue
		}
		seenTickers[currency.Ticker] = true

		var totalBalance decimal.Decimal

		if tradingBalance, exists := tradingBalanceMap[currency.Ticker]; exists {
			totalBalance = totalBalance.Add(tradingBalance)
		}

		if fundingBalance, exists := fundingBalanceMap[currency.Ticker]; exists {
			totalBalance = totalBalance.Add(fundingBalance)
		}

		if !totalBalance.IsZero() {
			amountUSD, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
				Source:     models.ExchangeSlugBybit.String(),
				From:       currency.Ticker,
				To:         models.CurrencyCodeUSDT,
				Amount:     totalBalance.String(),
				StableCoin: false,
			})
			if err != nil {
				continue
			}

			balances = append(balances, &models.AccountBalanceDTO{
				Currency:  currency.ID.String,
				Type:      models.CurrencyTypeCrypto.String(),
				Amount:    totalBalance,
				AmountUSD: amountUSD.Round(4),
			})
		}
	}

	return balances, nil
}

func (o *Service) GetCurrencyBalance(ctx context.Context, currency string) (*decimal.Decimal, error) {
	tradingBalances, err := o.exClient.Account().GetTradingBalance(ctx, &requests.GetTradingBalanceRequest{
		AccountType: bybitmodels.AccountTypeUnified.String(),
		Coin:        currency,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get trading balance: %w", err)
	}

	fundingBalances, err := o.exClient.Account().GetAllCoinsBalance(ctx, &requests.GetFundingBalanceRequest{
		AccountType: bybitmodels.AccountTypeFund.String(),
		Coin:        currency,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get funding balance: %w", err)
	}

	tradingAmt, fundingAmt := decimal.Zero, decimal.Zero

	if len(tradingBalances.Result.List) > 0 && len(tradingBalances.Result.List[0].Coin) > 0 {
		for _, coin := range tradingBalances.Result.List[0].Coin {
			if coin.Coin == currency {
				if coin.WalletBalance == "" {
					tradingAmt = decimal.Zero
				} else {
					tradingAmt, _ = decimal.NewFromString(coin.WalletBalance)
				}
				break
			}
		}
	}

	if len(fundingBalances.Result.Balance) > 0 {
		for _, balance := range fundingBalances.Result.Balance {
			if balance.Coin == currency {
				if balance.WalletBalance == "" {
					fundingAmt = decimal.Zero
				} else {
					fundingAmt, _ = decimal.NewFromString(balance.WalletBalance)
				}
				break
			}
		}
	}

	amt := tradingAmt.Add(fundingAmt)
	return &amt, nil
}

func (o *Service) GetExchangeSymbols(ctx context.Context) ([]*models.ExchangeSymbolDTO, error) {
	res, err := o.exClient.Market().GetInstruments(ctx, &requests.GetInstrumentsRequest{
		Category: "spot",
	})
	if err != nil {
		return nil, err
	}

	dto := make([]*models.ExchangeSymbolDTO, 0, len(res.Result.List)*2)
	for _, s := range res.Result.List {
		if s.Status != bybitmodels.InstrumentStatusStatusTrading {
			continue
		}
		base, quote := strings.ToUpper(s.BaseCoin), strings.ToUpper(s.QuoteCoin)
		dto = append(dto, &models.ExchangeSymbolDTO{
			Symbol:      s.Symbol,
			DisplayName: base + "/" + quote,
			BaseSymbol:  s.BaseCoin,
			QuoteSymbol: s.QuoteCoin,
			Type:        "sell",
		}, &models.ExchangeSymbolDTO{
			Symbol:      s.Symbol,
			DisplayName: quote + "/" + base,
			BaseSymbol:  s.BaseCoin,
			QuoteSymbol: s.QuoteCoin,
			Type:        "buy",
		})
	}

	return dto, nil
}

func (o *Service) GetDepositAddresses(ctx context.Context, currency, chain string) ([]*models.DepositAddressDTO, error) {
	enabledCurrencies, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugBybit)
	if err != nil {
		return nil, fmt.Errorf("fetch enabled currencies: %w", err)
	}

	filteredCurrencies := lo.Filter(enabledCurrencies, func(i *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) bool {
		if chain != "" {
			return i.Ticker == currency && i.Chain == chain
		}
		return i.Ticker == currency
	})

	if len(filteredCurrencies) == 0 {
		return nil, fmt.Errorf("currency %s with chain %s not found in enabled currencies", currency, chain)
	}

	res, err := o.exClient.Account().GetDepositAddress(ctx, &requests.GetDepositAddressRequest{
		Coin:      currency,
		ChainType: chain, // Optional parameter - if empty, gets all chains
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get deposit address: %w", err)
	}

	deposits := make([]*models.DepositAddressDTO, 0, len(filteredCurrencies))

	for _, chainInfo := range res.Result.Chains {
		matchingCurrency, found := lo.Find(filteredCurrencies, func(i *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
			return i.Chain == chainInfo.Chain
		})
		if !found {
			continue
		}

		currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByParams(ctx, repo_exchange_chains.GetCurrencyIDByParamsParams{
			Ticker: matchingCurrency.Ticker,
			Chain:  matchingCurrency.Chain,
			Slug:   models.ExchangeSlugBybit,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get internal currency id %s: %w", matchingCurrency.Chain, err)
		}

		deposits = append(deposits, &models.DepositAddressDTO{
			Address:          chainInfo.AddressDeposit,
			Currency:         currencyID,
			Chain:            chainInfo.Chain,
			AddressType:      models.DepositAddress,
			InternalCurrency: res.Result.Coin,
			PaymentTag:       chainInfo.TagDeposit, // Optional field, for TON/XRP and similar chains
		})
	}

	return deposits, nil
}

func (o *Service) GetWithdrawalRules(ctx context.Context, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	currEnabled, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugBybit)
	if err != nil {
		return nil, err
	}

	currEnabled = lo.Filter(currEnabled, func(i *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) bool {
		return lo.Contains(currencies, i.ID.String)
	})

	ccys := lo.Map(currEnabled, func(i *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) string {
		return i.Ticker
	})

	currReferences := make([]bybitmodels.CoinInfo, 0, len(ccys))
	for _, currency := range ccys {
		res, err := o.exClient.Market().GetCoinInfo(ctx, &requests.GetCoinInfoRequest{
			Coin: currency,
		})
		if err != nil {
			return nil, err
		}
		currReferences = append(currReferences, res.Result.Rows...)
	}

	exchangeRules := make([]*models.WithdrawalRulesDTO, 0, len(currReferences))
	for _, item := range currReferences {
		if slices.ContainsFunc(currEnabled, func(c *repo_exchange_chains.GetEnabledCurrenciesRow) bool { //nolint:nestif
			return c.Ticker == item.Coin
		}) {
			for _, network := range item.Chains {
				if slices.ContainsFunc(currEnabled, func(c *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
					return c.Chain == network.Chain && item.Coin == c.Ticker
				}) {
					feeType := models.WithdrawalFeeTypeFixed
					feeAmount := network.WithdrawFee

					rule := &models.WithdrawalRulesDTO{
						Currency:           item.Coin,
						Chain:              network.Chain,
						MinWithdrawAmount:  network.WithdrawMin,
						NumOfConfirmations: network.Confirmation,
						MinDepositAmount:   network.DepositMin,
						WithdrawPrecision:  network.MinAccuracy,
						Fee:                feeAmount,
						WithdrawFeeType:    feeType,
					}

					if network.DepositMin == "0" || network.DepositMin == "" {
						// If deposit minimum is zero or empty, set it to the 1 USDT equivalent
						minDepositAmount, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
							Source:     models.ExchangeSlugBybit.String(),
							From:       models.CurrencyCodeUSDT,
							To:         item.Coin,
							Amount:     "1",
							StableCoin: false,
						})
						if err != nil {
							return nil, fmt.Errorf("convert 1 USDT to %s: %w", item.Coin, err)
						}
						if minDepositAmount.LessThan(decimal.NewFromFloat(1.1)) && minDepositAmount.GreaterThan(decimal.NewFromFloat(0.9)) {
							// If the converted amount is around 1, set it to 1
							minDepositAmount = decimal.NewFromFloat(1)
						} else {
							precision := int32(6) // Default precision for crypto amounts
							if network.MinAccuracy != "" {
								if parsedPrecision, err := strconv.Atoi(network.MinAccuracy); err == nil {
									precision = int32(parsedPrecision) //nolint:gosec
								}
							}
							minDepositAmount = minDepositAmount.RoundUp(precision)
						}
						rule.MinDepositAmount = minDepositAmount.String()
					}

					exchangeRules = append(exchangeRules, rule)
				}
			}
		}
	}

	return exchangeRules, nil
}

func (o *Service) GetOrderRule(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	res, err := o.exClient.Market().GetInstruments(ctx, &requests.GetInstrumentsRequest{
		Category: "spot",
		Symbol:   ticker,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get instrument info: %w", err)
	}

	if len(res.Result.List) == 0 {
		return nil, fmt.Errorf("symbol %s not found", ticker)
	}

	symbolData := res.Result.List[0]

	// Check if symbol is available for trading
	if symbolData.Status != bybitmodels.InstrumentStatusStatusTrading {
		return nil, fmt.Errorf("symbol %s is not available for trading: status=%s", ticker, symbolData.Status)
	}

	// Calculate precision from string representations
	basePrecision := utils.ConvertPrecision(symbolData.LotSizeFilter.BasePrecision)
	quotePrecision := utils.ConvertPrecision(symbolData.LotSizeFilter.QuotePrecision)
	pricePrecision := utils.ConvertPrecision(symbolData.PriceFilter.TickSize)

	minOrderAmount := symbolData.LotSizeFilter.MinOrderQty
	maxOrderAmount := symbolData.LotSizeFilter.MaxOrderQty
	minOrderValue := symbolData.LotSizeFilter.MinOrderAmt

	// Handle currency conversions for minimum amounts
	adjustedMinOrderValue := minOrderValue
	adjustedMinOrderAmount := minOrderAmount

	// Get current price of the trading pair to calculate minimum amounts
	tickerData, err := o.exClient.Market().GetTickers(ctx, &requests.GetTickersRequest{
		Category: "spot",
		Symbol:   ticker,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ticker data: %w", err)
	}

	if len(tickerData.Result.List) == 0 {
		return nil, fmt.Errorf("ticker data for %s not found", ticker)
	}

	currentPrice, err := decimal.NewFromString(tickerData.Result.List[0].LastPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current price: %w", err)
	}

	// Calculate minimum amount based on minimum notional value and current price
	// minOrderValue is in quote currency, divide by price to get base currency amount
	calculatedMinAmount := minOrderValue.Div(currentPrice).RoundUp(int32(basePrecision.IntPart())) //nolint:gosec

	// Use the larger of the calculated minimum or the exchange's minimum quantity
	if calculatedMinAmount.GreaterThan(minOrderAmount) {
		adjustedMinOrderAmount = calculatedMinAmount
	} else {
		adjustedMinOrderAmount = minOrderAmount
	}
	dto := &models.OrderRulesDTO{
		Symbol:          symbolData.Symbol,
		State:           symbolData.Status.String(),
		BaseCurrency:    symbolData.BaseCoin,
		QuoteCurrency:   symbolData.QuoteCoin,
		PricePrecision:  int(pricePrecision.IntPart()),
		AmountPrecision: int(basePrecision.IntPart()),
		ValuePrecision:  int(quotePrecision.IntPart()),
		MinOrderAmount:  adjustedMinOrderAmount.String(),
		MaxOrderAmount:  maxOrderAmount.String(),
		MinOrderValue:   adjustedMinOrderValue.String(),
		// Bybit uses the same max values for market orders in spot
		SellMarketMinOrderAmount: adjustedMinOrderAmount.String(),
		SellMarketMaxOrderAmount: maxOrderAmount.String(),
		BuyMarketMaxOrderValue:   symbolData.LotSizeFilter.MaxOrderAmt.String(),
	}

	return dto, nil
}

func (o *Service) GetOrderRules(ctx context.Context, ticker ...string) ([]*models.OrderRulesDTO, error) {
	if len(ticker) == 0 {
		return nil, fmt.Errorf("ticker is required")
	}
	rules := make([]*models.OrderRulesDTO, 0, len(ticker))
	for _, t := range ticker {
		rule, err := o.GetOrderRule(ctx, t)
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (o *Service) GetOrderDetails(ctx context.Context, args *models.GetOrderByIDParams) (*models.OrderDetailsDTO, error) { //nolint:gocognit,gocyclo,funlen
	order := &models.OrderDetailsDTO{
		State:     models.ExchangeOrderStatusFailed,
		Amount:    decimal.Zero,
		AmountUSD: decimal.Zero,
	}

	if args.ExternalOrderID != nil { //nolint:nestif
		// Query the specific order by ID
		res, err := o.exClient.Trade().GetActiveOrders(ctx, &requests.GetActiveOrdersRequest{
			Category: "spot",
			OrderID:  *args.ExternalOrderID,
		})
		if err != nil {
			return nil, fmt.Errorf("get spot order %s: %w", *args.ExternalOrderID, err)
		}

		if len(res.Result.List) == 0 {
			// Try order history for older orders
			historyRes, err := o.exClient.Trade().GetOrderHistory(ctx, &requests.GetOrderHistoryRequest{
				Category: "spot",
				OrderID:  *args.ExternalOrderID,
			})
			if err != nil {
				return nil, fmt.Errorf("get order history %s: %w", *args.ExternalOrderID, err)
			}

			if len(historyRes.Result.List) == 0 {
				return order, nil // No order found in history either, leave it as failed
			}

			// Use the historical order data
			res.Result.List = historyRes.Result.List
		}

		orderData := res.Result.List[0]

		// Map order status
		switch orderData.OrderStatus {
		case bybitmodels.OrderStatusFilled:
			order.State = models.ExchangeOrderStatusCompleted
		case bybitmodels.OrderStatusNew, bybitmodels.OrderStatusPartiallyFilled:
			order.State = models.ExchangeOrderStatusInProgress
		case bybitmodels.OrderStatusCancelled,
			bybitmodels.OrderStatusRejected,
			bybitmodels.OrderStatusDeactivated:
			order.State = models.ExchangeOrderStatusFailed
		default:
			order.State = models.ExchangeOrderStatusInProgress
		}

		// Get symbol information
		symbolInfo, err := o.exClient.Market().GetInstruments(ctx, &requests.GetInstrumentsRequest{
			Category: "spot",
			Symbol:   orderData.Symbol,
		})
		if err != nil {
			return nil, fmt.Errorf("get trading symbol %s: %w", orderData.Symbol, err)
		}

		if len(symbolInfo.Result.List) == 0 {
			return nil, fmt.Errorf("symbol %s not found", orderData.Symbol)
		}
		symbol := symbolInfo.Result.List[0]

		// Get current ticker for price information
		ticker, err := o.exClient.Market().GetTickers(ctx, &requests.GetTickersRequest{
			Category: "spot",
			Symbol:   orderData.Symbol,
		})
		if err != nil {
			return nil, fmt.Errorf("get ticker info for %s: %w", orderData.Symbol, err)
		}

		if len(ticker.Result.List) == 0 {
			return nil, fmt.Errorf("ticker data for %s not found", orderData.Symbol)
		}

		// Parse executed quantity (always in base currency for Bybit)
		executedQty, err := decimal.NewFromString(orderData.CumExecQty)
		if err != nil {
			return nil, fmt.Errorf("parse executed quantity %s: %w", orderData.CumExecQty, err)
		}

		order.Amount = executedQty

		// Calculate USD value based on the trading pair
		currentPrice := ticker.Result.List[0].LastPrice
		price, err := decimal.NewFromString(currentPrice)
		if err != nil {
			return nil, fmt.Errorf("parse current price %s: %w", currentPrice, err)
		}

		switch {
		case symbol.QuoteCoin == models.CurrencyCodeUSDT || symbol.QuoteCoin == models.CurrencyCodeUSDC:
			// Quote is already USD (e.g., BTC/USDT)
			// USD value = base amount * price
			order.AmountUSD = order.Amount.Mul(price).Round(4)

		case symbol.BaseCoin == models.CurrencyCodeUSDT || symbol.BaseCoin == models.CurrencyCodeUSDC:
			// Base is USD (e.g., USDT/EUR) - rare case
			order.AmountUSD = order.Amount

		default:
			// Neither base nor quote is USD (e.g., BTC/ETH)
			// Need to convert base to USD
			baseUSDSymbol := symbol.BaseCoin + "USDT"
			baseUSDTicker, err := o.exClient.Market().GetTickers(ctx, &requests.GetTickersRequest{
				Category: "spot",
				Symbol:   baseUSDSymbol,
			})
			if err != nil {
				// Try USDC if USDT fails
				baseUSDSymbol = symbol.BaseCoin + "USDC"
				baseUSDTicker, err = o.exClient.Market().GetTickers(ctx, &requests.GetTickersRequest{
					Category: "spot",
					Symbol:   baseUSDSymbol,
				})
				if err != nil {
					return nil, fmt.Errorf("get USD ticker for %s: %w", symbol.BaseCoin, err)
				}
			}

			if len(baseUSDTicker.Result.List) > 0 {
				baseUSDPrice, err := decimal.NewFromString(baseUSDTicker.Result.List[0].LastPrice)
				if err != nil {
					return nil, fmt.Errorf("parse base USD price: %w", err)
				}
				// USD value = amount of BASE * price of BASE in USD
				order.AmountUSD = order.Amount.Mul(baseUSDPrice).Round(4)
			}
		}

		return order, nil
	}

	if args.ClientOrderID != nil { //nolint:nestif
		// Query by client order ID (orderLinkId) - same logic as ExternalOrderID
		res, err := o.exClient.Trade().GetActiveOrders(ctx, &requests.GetActiveOrdersRequest{
			Category:    "spot",
			OrderLinkID: *args.ClientOrderID,
		})
		if err != nil {
			return nil, fmt.Errorf("get spot order by client ID %s: %w", *args.ClientOrderID, err)
		}

		if len(res.Result.List) == 0 {
			// Try order history for older orders
			historyRes, err := o.exClient.Trade().GetOrderHistory(ctx, &requests.GetOrderHistoryRequest{
				Category:    "spot",
				OrderLinkID: *args.ClientOrderID,
			})
			if err != nil {
				return nil, fmt.Errorf("get order history by client ID %s: %w", *args.ClientOrderID, err)
			}

			if len(historyRes.Result.List) == 0 {
				return order, nil // No order found in history either, leave it as failed
			}

			// Use the historical order data
			res.Result.List = historyRes.Result.List
		}

		orderData := res.Result.List[0]

		switch orderData.OrderStatus {
		case bybitmodels.OrderStatusFilled:
			order.State = models.ExchangeOrderStatusCompleted
		case bybitmodels.OrderStatusNew, bybitmodels.OrderStatusPartiallyFilled:
			order.State = models.ExchangeOrderStatusInProgress
		case bybitmodels.OrderStatusCancelled,
			bybitmodels.OrderStatusRejected,
			bybitmodels.OrderStatusDeactivated:
			order.State = models.ExchangeOrderStatusFailed
		default:
			order.State = models.ExchangeOrderStatusInProgress
		}

		// Parse executed quantity
		executedQty, err := decimal.NewFromString(orderData.CumExecQty)
		if err != nil {
			return nil, fmt.Errorf("parse executed quantity %s: %w", orderData.CumExecQty, err)
		}
		order.Amount = executedQty

		// Get symbol information for USD calculation
		symbolInfo, err := o.exClient.Market().GetInstruments(ctx, &requests.GetInstrumentsRequest{
			Category: "spot",
			Symbol:   orderData.Symbol,
		})
		if err != nil {
			return nil, fmt.Errorf("get trading symbol %s: %w", orderData.Symbol, err)
		}

		if len(symbolInfo.Result.List) > 0 {
			symbol := symbolInfo.Result.List[0]

			// Calculate USD value (same logic as ExternalOrderID case)
			switch {
			case symbol.QuoteCoin == models.CurrencyCodeUSDT || symbol.QuoteCoin == models.CurrencyCodeUSDC:
				ticker, err := o.exClient.Market().GetTickers(ctx, &requests.GetTickersRequest{
					Category: "spot",
					Symbol:   orderData.Symbol,
				})
				if err == nil && len(ticker.Result.List) > 0 {
					if price, err := decimal.NewFromString(ticker.Result.List[0].LastPrice); err == nil {
						order.AmountUSD = order.Amount.Mul(price).Round(4)
					}
				}
			case symbol.BaseCoin == models.CurrencyCodeUSDT || symbol.BaseCoin == models.CurrencyCodeUSDC:
				order.AmountUSD = order.Amount
			default:
				// Try to get base currency USD price
				baseUSDSymbol := symbol.BaseCoin + "USDT"
				if baseUSDTicker, err := o.exClient.Market().GetTickers(ctx, &requests.GetTickersRequest{
					Category: "spot",
					Symbol:   baseUSDSymbol,
				}); err == nil && len(baseUSDTicker.Result.List) > 0 {
					if baseUSDPrice, err := decimal.NewFromString(baseUSDTicker.Result.List[0].LastPrice); err == nil {
						order.AmountUSD = order.Amount.Mul(baseUSDPrice).Round(4)
					}
				}
			}
		}

		return order, nil
	}

	return order, nil
}

func (o *Service) CreateSpotOrder(ctx context.Context, from string, to string, side string, ticker string, amount *decimal.Decimal, rule *models.OrderRulesDTO) (*models.ExchangeOrderDTO, error) { //nolint:all
	clientOrderID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	// Get instrument information
	res, err := o.exClient.Market().GetInstruments(ctx, &requests.GetInstrumentsRequest{
		Category: "spot",
		Symbol:   ticker,
	})
	if err != nil {
		return nil, fmt.Errorf("get trading symbol %s: %w", ticker, err)
	}
	if len(res.Result.List) == 0 {
		return nil, fmt.Errorf("symbol %s not found", ticker)
	}

	symbol := res.Result.List[0]
	if symbol.Status != bybitmodels.InstrumentStatusStatusTrading {
		return nil, fmt.Errorf("symbol %s is not available for trading: status=%s", ticker, symbol.Status)
	}

	// Prepare order request
	spotOrderRequest := &requests.PlaceOrderRequest{
		Category:    "spot",
		Symbol:      symbol.Symbol,
		OrderLinkID: clientOrderID.String(),
		OrderType:   bybitmodels.OrderTypeMarket.String(),
	}

	// explicitly set the side to lowercase
	switch side {
	case "sell":
		spotOrderRequest.Side = bybitmodels.SideSell.String()
	case "buy":
		spotOrderRequest.Side = bybitmodels.SideBuy.String()
	default:
		return nil, fmt.Errorf("unsupported order side %s", side)
	}

	var unifiedBaseBalance, unifiedQuoteBalance, fundingBaseBalance, fundingQuoteBalance decimal.Decimal
	{
		res, err := o.exClient.Account().GetTradingBalance(ctx, &requests.GetTradingBalanceRequest{
			AccountType: bybitmodels.AccountTypeUnified.String(),
			Coin:        symbol.BaseCoin,
		})
		if err != nil {
			return nil, fmt.Errorf("get unified base currency balance %s: %w", symbol.BaseCoin, err)
		}
		if len(res.Result.List) > 0 && len(res.Result.List[0].Coin) > 0 {
			if res.Result.List[0].Coin[0].WalletBalance == "" {
				unifiedBaseBalance = decimal.Zero
			} else {
				unifiedBaseBalance, _ = decimal.NewFromString(res.Result.List[0].Coin[0].WalletBalance)
			}
		}
	}
	{
		res, err := o.exClient.Account().GetTradingBalance(ctx, &requests.GetTradingBalanceRequest{
			AccountType: bybitmodels.AccountTypeUnified.String(),
			Coin:        symbol.QuoteCoin,
		})
		if err != nil {
			return nil, fmt.Errorf("get unified quote currency balance %s: %w", symbol.QuoteCoin, err)
		}
		if len(res.Result.List) > 0 && len(res.Result.List[0].Coin) > 0 {
			if res.Result.List[0].Coin[0].WalletBalance == "" {
				unifiedQuoteBalance = decimal.Zero
			} else {
				unifiedQuoteBalance, _ = decimal.NewFromString(res.Result.List[0].Coin[0].WalletBalance)
			}
		}
	}
	{
		res, err := o.exClient.Account().GetAllCoinsBalance(ctx, &requests.GetFundingBalanceRequest{
			AccountType: bybitmodels.AccountTypeFund.String(),
			Coin:        symbol.BaseCoin,
		})
		if err != nil {
			return nil, fmt.Errorf("get funding base currency balance %s: %w", symbol.BaseCoin, err)
		}
		if len(res.Result.Balance) > 0 && len(res.Result.Balance[0].Coin) > 0 {
			if res.Result.Balance[0].WalletBalance == "" {
				fundingBaseBalance = decimal.Zero
			} else {
				fundingBaseBalance, _ = decimal.NewFromString(res.Result.Balance[0].WalletBalance)
			}
		}
	}
	{
		res, err := o.exClient.Account().GetAllCoinsBalance(ctx, &requests.GetFundingBalanceRequest{
			AccountType: bybitmodels.AccountTypeFund.String(),
			Coin:        symbol.QuoteCoin,
		})
		if err != nil {
			return nil, fmt.Errorf("get funding quote currency balance %s: %w", symbol.QuoteCoin, err)
		}
		if len(res.Result.Balance) > 0 && len(res.Result.Balance[0].Coin) > 0 {
			if res.Result.Balance[0].WalletBalance == "" {
				fundingQuoteBalance = decimal.Zero
			} else {
				fundingQuoteBalance, _ = decimal.NewFromString(res.Result.Balance[0].WalletBalance)
			}
		}
	}

	// Parse minimum order requirements
	orderMinimumBase, err := decimal.NewFromString(rule.MinOrderAmount)
	if err != nil {
		return nil, fmt.Errorf("failed to parse minimum order amount: %w", err)
	}

	orderMinimumQuote, err := decimal.NewFromString(rule.MinOrderValue)
	if err != nil {
		return nil, fmt.Errorf("failed to parse minimum order value: %w", err)
	}

	// Determine which currency and accounts to use based on order side
	var transferCurrency string
	var unifiedBalance, fundingBalance, totalBalance, minimumRequired decimal.Decimal

	switch spotOrderRequest.Side {
	case bybitmodels.SideSell.String():
		transferCurrency = symbol.BaseCoin
		unifiedBalance = unifiedBaseBalance
		fundingBalance = fundingBaseBalance
		totalBalance = unifiedBalance.Add(fundingBalance)
		minimumRequired = orderMinimumBase

		if totalBalance.LessThan(minimumRequired) {
			return nil, ErrInsufficientBalance
		}
	case bybitmodels.SideBuy.String():
		transferCurrency = symbol.QuoteCoin
		unifiedBalance = unifiedQuoteBalance
		fundingBalance = fundingQuoteBalance
		totalBalance = unifiedBalance.Add(fundingBalance)
		minimumRequired = orderMinimumQuote

		if totalBalance.LessThan(minimumRequired) {
			return nil, ErrInsufficientBalance
		}
	default:
		return nil, fmt.Errorf("unsupported order side %s", spotOrderRequest.Side)
	}

	// Transfer funds from funding to unified if needed
	remainingTopup := totalBalance.Sub(unifiedBalance)
	if remainingTopup.GreaterThan(decimal.Zero) && fundingBalance.GreaterThan(decimal.Zero) {
		transferAmount := remainingTopup
		if transferAmount.GreaterThan(fundingBalance) {
			transferAmount = fundingBalance
		}

		transferID, err := uuid.NewUUID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate transfer ID: %w", err)
		}

		transferReq := &requests.CreateInternalTransferRequest{
			TransferID:      transferID.String(),
			Coin:            transferCurrency,
			Amount:          transferAmount.String(),
			FromAccountType: bybitmodels.AccountTypeFund,
			ToAccountType:   bybitmodels.AccountTypeUnified,
		}

		_, err = o.exClient.Account().CreateInternalTransfer(ctx, transferReq)
		if err != nil {
			return nil, fmt.Errorf("failed to transfer funds from funding to unified: %w", err)
		}
	}

	// Set order quantity based on side
	maxAmount := totalBalance
	switch spotOrderRequest.Side {
	case bybitmodels.SideSell.String():
		spotOrderRequest.Qty = maxAmount.RoundDown(int32(rule.AmountPrecision)).String()
		// MarketUnit defaults to baseCoin for sell orders, so we don't need to set it
	case bybitmodels.SideBuy.String():
		spotOrderRequest.Qty = maxAmount.RoundDown(int32(rule.ValuePrecision)).String()
		spotOrderRequest.MarketUnit = bybitmodels.MarketUnitQuoteCoin.String()
	}

	placedOrder, err := o.exClient.Trade().PlaceOrder(ctx, spotOrderRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	// Update amount to reflect what was actually used
	amount = &maxAmount //nolint:staticcheck

	return &models.ExchangeOrderDTO{
		ClientOrderID:   placedOrder.Result.OrderLinkID,
		ExchangeOrderID: placedOrder.Result.OrderID,
		Amount:          *amount,
	}, nil
}

func (o *Service) CreateWithdrawalOrder(ctx context.Context, args *models.CreateWithdrawalOrderParams) (*models.ExchangeWithdrawalDTO, error) {
	precision := int32(args.WithdrawalPrecision)

	args.NativeAmount = args.NativeAmount.RoundDown(precision)

	internalCurrencyID, err := o.storage.ExchangeChains().GetTickerByCurrencyID(ctx, repo_exchange_chains.GetTickerByCurrencyIDParams{
		CurrencyID: args.Currency,
		Slug:       models.ExchangeSlugBybit,
	})
	if err != nil {
		return nil, err
	}

	// Check FUND account balance
	fundingBalance, err := o.exClient.Account().GetAllCoinsBalance(ctx, &requests.GetFundingBalanceRequest{
		AccountType: bybitmodels.AccountTypeFund.String(),
		Coin:        internalCurrencyID,
	})
	if err != nil {
		return nil, err
	}

	// Check UNIFIED account balance
	unifiedBalance, err := o.exClient.Account().GetAllCoinsBalance(ctx, &requests.GetFundingBalanceRequest{
		AccountType: bybitmodels.AccountTypeUnified.String(),
		Coin:        internalCurrencyID,
	})
	if err != nil {
		return nil, err
	}

	fundingAmount := o.getBalanceByCurrency(internalCurrencyID, fundingBalance.Result.Balance).RoundDown(precision)
	unifiedAmount := o.getBalanceByCurrency(internalCurrencyID, unifiedBalance.Result.Balance).RoundDown(precision)

	totalBalance := fundingAmount.Add(unifiedAmount)

	o.l.Infow("balances",
		"exchange", models.ExchangeSlugBybit.String(),
		"recordID", args.RecordID.String(),
		"withdrawalAmount", args.NativeAmount.String(),
		"withdrawalFee", args.Fee.String(),
		"totalBalance", totalBalance.String(),
		"fundingBalance", fundingAmount.String(),
		"unifiedBalance", unifiedAmount.String(),
	)

	// Bybit requires withdrawals to be made from FUND account
	// If insufficient funds in FUND, transfer from UNIFIED
	if fundingAmount.LessThan(args.NativeAmount) {
		o.l.Infow("funding balance is less than withdrawal amount",
			"exchange", models.ExchangeSlugBybit.String(),
			"recordID", args.RecordID.String(),
		)

		if totalBalance.LessThan(args.NativeAmount) {
			o.l.Infow("total balance is less than withdrawal amount",
				"exchange", models.ExchangeSlugBybit.String(),
				"recordID", args.RecordID.String(),
			)
			return nil, exchangeclient.ErrMinWithdrawalBalance
		}

		transferAmount := args.NativeAmount.Sub(fundingAmount).RoundDown(precision)

		transferID, err := uuid.NewUUID()
		if err != nil {
			return nil, err
		}

		_, err = o.exClient.Account().CreateInternalTransfer(ctx, &requests.CreateInternalTransferRequest{
			TransferID:      transferID.String(),
			Coin:            internalCurrencyID,
			Amount:          transferAmount.String(),
			FromAccountType: bybitmodels.AccountTypeUnified,
			ToAccountType:   bybitmodels.AccountTypeFund,
		})
		if err != nil {
			return nil, err
		}

		o.l.Infow("transferred funds from UNIFIED to FUND",
			"exchange", models.ExchangeSlugBybit.String(),
			"recordID", args.RecordID.String(),
			"transferAmount", transferAmount.String(),
		)
	}

	req := &requests.CreateWithdrawRequest{
		Coin:      internalCurrencyID,
		Chain:     args.Chain,
		Address:   args.Address,
		Amount:    args.NativeAmount.Sub(args.Fee).RoundDown(precision).String(),
		Timestamp: strconv.Itoa(int(time.Now().UnixMilli())),
		FeeType:   "1", // 0: return fee to wallet, 1: deduct fee from withdrawal amount
	}

	o.l.Infow("withdrawal request assembled",
		"exchange", models.ExchangeSlugBybit.String(),
		"recordID", args.RecordID.String(),
		"request", req,
		"amount", req.Amount,
		"currency", req.Coin,
		"chain", req.Chain,
		"address", req.Address,
	)

	amount := args.NativeAmount.Sub(args.Fee)
	minWithdrawal := args.MinWithdrawal

	dto := &models.ExchangeWithdrawalDTO{}

	for {
		if amount.LessThan(minWithdrawal) {
			o.l.Infow("withdrawal amount below minimum",
				"exchange", models.ExchangeSlugBybit.String(),
				"recordID", args.RecordID.String(),
				"current_amount", amount.String(),
				"min_withdrawal", minWithdrawal.String(),
			)
			return nil, exchangeclient.ErrWithdrawalBalanceLocked
		}

		withdrawalStep, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
			Source:     models.ExchangeSlugBybit.String(),
			From:       "USDT",
			To:         req.Coin,
			Amount:     decimal.NewFromInt(WithdrawalStep).String(),
			StableCoin: false,
		})
		if err != nil {
			return nil, err
		}

		req.Amount = amount.String()
		res, err := o.exClient.Account().CreateWithdraw(ctx, req)
		if err == nil {
			dto.InternalOrderID = res.Result.ID
			dto.ExternalOrderID = res.Result.ID
			return dto, nil
		}

		// Check for Bybit-specific insufficient balance errors
		if errors.Is(err, exchangeclient.ErrWithdrawalBalanceLocked) {
			o.l.Errorw("insufficient funds, retrying with reduced amount",
				"error", exchangeclient.ErrWithdrawalBalanceLocked,
				"exchange", models.ExchangeSlugBybit.String(),
				"recordID", args.RecordID.String(),
				"current_amount", amount.String(),
			)

			amount = amount.Sub(withdrawalStep)
			if amount.LessThan(minWithdrawal) {
				return nil, exchangeclient.ErrMinWithdrawalBalance
			}
			dto.RetryReason = exchangeclient.ErrWithdrawalBalanceLocked.Error()
			continue
		}

		return nil, err
	}
}

func (o *Service) getBalanceByCurrency(currency string, balances []bybitmodels.AccountBalance) decimal.Decimal {
	for _, balance := range balances {
		if balance.Coin == currency {
			if balance.WalletBalance == "" {
				return decimal.Zero
			}
			bal, _ := decimal.NewFromString(balance.WalletBalance)
			return bal
		}
	}
	return decimal.Zero
}

func (o *Service) GetWithdrawalByID(ctx context.Context, args *models.GetWithdrawalByIDParams) (*models.WithdrawalStatusDTO, error) {
	request := &requests.GetWithdrawRequest{
		WithdrawType: "0",
	}

	if args.ClientOrderID == nil {
		return nil, fmt.Errorf("client order ID is required")
	}
	if args.ClientOrderID != nil {
		request.WithdrawID = *args.ClientOrderID
	}

	wdRecords, err := o.exClient.Account().GetWithdraw(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawal by ID %s: %w", *args.ClientOrderID, err)
	}
	if len(wdRecords.Result.Rows) == 0 {
		return nil, fmt.Errorf("withdrawal with ID %s not found", *args.ClientOrderID)
	}
	record := wdRecords.Result.Rows[0]

	res := &models.WithdrawalStatusDTO{
		ID:     record.WithdrawID,
		Status: record.Status.String(),
	}
	if record.TxID != "" {
		res.TxHash = record.TxID
	}
	if !record.Amount.IsZero() {
		res.NativeAmount = record.Amount
	}

	return res, nil
}

func (o *Service) GetConnectionHash() string { return o.connHash }
