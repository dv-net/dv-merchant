package gateio

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_chains"
	"github.com/dv-net/dv-merchant/internal/tools/hash"
	"github.com/dv-net/dv-merchant/pkg/iso"
	"github.com/dv-net/dv-merchant/pkg/logger"
	"github.com/dv-net/dv-merchant/pkg/retry"

	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/ulule/limiter/v3"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	gateio "github.com/dv-net/dv-merchant/pkg/exchange_client/gate"
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
	exClient *gateio.BaseClient
	storage  storage.IStorage
	convSvc  currconv.ICurrencyConvertor
	l        logger.Logger
	connHash string
}

func NewService(logger logger.Logger, accessKey, secretKey string, baseURL *url.URL, storage storage.IStorage, store limiter.Store, convSvc currconv.ICurrencyConvertor) (*Service, error) {
	exClient, err := gateio.NewBaseClient(&gateio.ClientOptions{
		AccessKey: accessKey,
		SecretKey: secretKey,
		BaseURL:   baseURL,
	}, store, gateio.WithLogger(logger))
	if err != nil {
		return nil, err
	}
	connHash, err := hash.SHA256ConnectionHash(models.ExchangeSlugGateio.String(), accessKey, secretKey)
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
	if _, err := o.exClient.Account().GetAccountDetail(ctx); err != nil {
		return err
	}
	return nil
}

func (o *Service) GetAccountBalance(ctx context.Context) ([]*models.AccountBalanceDTO, error) {
	balances, err := o.exClient.Spot().GetSpotAccountBalances(ctx, &gateio.GetSpotAccountBalancesRequest{})
	if err != nil {
		return nil, fmt.Errorf("get account balance: %w", err)
	}
	accountBalances := make([]*models.AccountBalanceDTO, 0, len(balances.Data))
	for _, balance := range balances.Data {
		currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByParams(ctx, repo_exchange_chains.GetCurrencyIDByParamsParams{
			Ticker: balance.Currency,
			Slug:   models.ExchangeSlugGateio,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		amountUSD, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
			Source: models.ExchangeSlugGateio.String(),
			From:   balance.Currency,
			To:     models.CurrencyCodeUSDT,
			Amount: balance.Available.String(),
		})
		if err != nil {
			return nil, err
		}

		accountBalances = append(accountBalances, &models.AccountBalanceDTO{
			Currency:  currencyID,
			Amount:    balance.Available,
			AmountUSD: amountUSD.Round(4),
			Type:      models.CurrencyTypeCrypto.String(),
		})
	}

	return accountBalances, nil
}

func (o *Service) GetCurrencyBalance(ctx context.Context, currency string) (*decimal.Decimal, error) {
	balance, err := o.exClient.Spot().GetSpotAccountBalances(ctx, &gateio.GetSpotAccountBalancesRequest{
		Currency: currency,
	})
	if err != nil {
		return nil, fmt.Errorf("get account balance: %w", err)
	}
	return &balance.Data[0].Available, nil
}

func (o *Service) GetExchangeSymbols(ctx context.Context) ([]*models.ExchangeSymbolDTO, error) {
	supportedPairs, err := o.exClient.Spot().GetSpotSupportedCurrencyPairs(ctx)
	if err != nil {
		return nil, fmt.Errorf("get supported currency pairs: %w", err)
	}

	symbols := make([]*models.ExchangeSymbolDTO, 0, len(supportedPairs.Data)*2)
	for _, pair := range supportedPairs.Data {
		if !gateio.PairTradeStatus(pair.TradeStatus).IsTradable() {
			continue
		}
		if iso.IsFiat(pair.Base) || iso.IsFiat(pair.Quote) {
			continue
		}
		if strings.HasSuffix(pair.Base, "3S") || strings.HasSuffix(pair.Quote, "3S") ||
			strings.HasSuffix(pair.Base, "3L") || strings.HasSuffix(pair.Quote, "3L") ||
			strings.HasSuffix(pair.Base, "5L") || strings.HasSuffix(pair.Quote, "5L") ||
			strings.HasSuffix(pair.Base, "5S") || strings.HasSuffix(pair.Quote, "5S") {
			continue
		}
		symbols = append(symbols, &models.ExchangeSymbolDTO{
			Symbol:      pair.ID,
			DisplayName: pair.Base + "/" + pair.Quote,
			BaseSymbol:  pair.Base,
			QuoteSymbol: pair.Quote,
			Type:        "sell",
		}, &models.ExchangeSymbolDTO{
			Symbol:      pair.ID,
			DisplayName: pair.Quote + "/" + pair.Base,
			BaseSymbol:  pair.Base,
			QuoteSymbol: pair.Quote,
			Type:        "buy",
		})
	}
	return symbols, nil
}

func (o *Service) GetDepositAddresses(ctx context.Context, currency, chain string) ([]*models.DepositAddressDTO, error) {
	var (
		exchangeAddresses *gateio.GetDepositAddressResponse
		err               error
	)
	retrier := retry.New(retry.WithContext(ctx), retry.WithDelay(time.Millisecond*500))
	err = retrier.Do(func() error {
		exchangeAddresses, err = o.exClient.Wallet().GetDepositAddress(ctx, &gateio.GetDepositAddressRequest{
			Currency: currency,
		})
		if err != nil {
			return fmt.Errorf("get deposit address for %s: %w", currency, err)
		}
		if strings.Contains(exchangeAddresses.Data.Address, "generated") {
			return fmt.Errorf("address is being generated for %s, retrying", currency)
		}
		if len(exchangeAddresses.Data.MultichainAddresses) == 0 {
			return fmt.Errorf("no multichain addresses found for %s", currency)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if exchangeAddresses.Data == nil || len(exchangeAddresses.Data.MultichainAddresses) == 0 {
		return nil, fmt.Errorf("no deposit addresses found for %s", currency)
	}

	addresses := make([]*models.DepositAddressDTO, 0, len(exchangeAddresses.Data.MultichainAddresses))
	for _, address := range exchangeAddresses.Data.MultichainAddresses {
		currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByParams(ctx, repo_exchange_chains.GetCurrencyIDByParamsParams{
			Ticker: currency,
			Chain:  address.Chain,
			Slug:   models.ExchangeSlugGateio,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get internal currency id %s: %w", address.Chain, err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if address.Chain == chain && exchangeAddresses.Data.Currency == currency {
			addresses = append(addresses, &models.DepositAddressDTO{
				Address:          address.Address,
				Currency:         currencyID,
				Chain:            address.Chain,
				AddressType:      models.DepositAddress,
				InternalCurrency: exchangeAddresses.Data.Currency,
			})
		}
	}
	return addresses, nil
}

func (o *Service) GetOrderRule(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	symbolData, err := o.exClient.Spot().GetSpotSupportedCurrencyPair(ctx, ticker)
	if err != nil {
		return nil, fmt.Errorf("get order rule for %s: %w", ticker, err)
	}
	if symbolData.Data == nil {
		return nil, fmt.Errorf("no data found for ticker %s", ticker)
	}

	minOrderAmount, err := decimal.NewFromString(symbolData.Data.MinBaseAmount)
	if err != nil {
		return nil, fmt.Errorf("parse min order amount for %s: %w", ticker, err)
	}
	minOrderValue, err := decimal.NewFromString(symbolData.Data.MinQuoteAmount)
	if err != nil {
		return nil, fmt.Errorf("parse min order value for %s: %w", ticker, err)
	}

	return &models.OrderRulesDTO{
		Symbol:          symbolData.Data.ID,
		BaseCurrency:    symbolData.Data.Base,
		QuoteCurrency:   symbolData.Data.Quote,
		MinOrderAmount:  minOrderAmount.String(),
		MinOrderValue:   minOrderValue.String(),
		PricePrecision:  symbolData.Data.Precision,
		AmountPrecision: symbolData.Data.AmountPrecision,
	}, nil
}

func (o *Service) GetOrderRules(ctx context.Context, tickers ...string) ([]*models.OrderRulesDTO, error) {
	rules := make([]*models.OrderRulesDTO, 0, len(tickers))
	for _, ticker := range tickers {
		rule, err := o.GetOrderRule(ctx, ticker)
		if err != nil {
			return nil, fmt.Errorf("get order rule for %s: %w", ticker, err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (o *Service) GetWithdrawalRules(ctx context.Context, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	currEnabled, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugGateio)
	if err != nil {
		return nil, fmt.Errorf("get enabled currencies: %w", err)
	}

	currEnabled = lo.Filter(currEnabled, func(item *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) bool {
		return lo.Contains(currencies, item.ID.String)
	})

	withdrawalRules := make([]*models.WithdrawalRulesDTO, 0, len(currEnabled))
	for _, currency := range currEnabled {
		currencyRules, err := o.exClient.Wallet().GetWithdrawalRules(ctx, &gateio.GetWithdrawalRulesRequest{
			Currency: currency.Ticker,
		})
		if err != nil {
			return nil, fmt.Errorf("get withdrawal rules for %s: %w", currency.Ticker, err)
		}
		if len(currencyRules.Data) != 0 { //nolint:nestif
			for chain, fee := range currencyRules.Data[0].WithdrawFixOnChains {
				if lo.ContainsBy(currEnabled, func(item *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
					return item.Chain == chain && item.Ticker == currency.Ticker
				}) {
					// Retrieve withdrawal precision
					chainInfo, err := o.exClient.Wallet().GetCurrencyChainsSupported(ctx, &gateio.GetCurrencyChainsSupportedRequest{
						Currency: currency.Ticker,
					})
					if err != nil {
						return nil, fmt.Errorf("get chains supported for currency %s: %w", currency.Ticker, err)
					}
					if len(chainInfo.Data) == 0 {
						return nil, fmt.Errorf("no chains found for currency %s", currency.Ticker)
					}
					withdrawalPrecision := decimal.NewFromInt(chainInfo.Data[0].Decimal)
					// Calculate minimum deposit amount as equivalent of one usdt
					minDepositAmount, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
						Source:     models.ExchangeSlugGateio.String(),
						From:       models.CurrencyCodeUSDT,
						To:         currency.Ticker,
						Amount:     "1",
						StableCoin: false,
					})
					if err != nil {
						return nil, fmt.Errorf("convert minimum deposit amount: %w", err)
					}
					withdrawalRules = append(withdrawalRules, &models.WithdrawalRulesDTO{
						Currency:           currency.Ticker,
						Chain:              chain,
						MinDepositAmount:   minDepositAmount.String(),
						MinWithdrawAmount:  currencyRules.Data[0].WithdrawAmountMini.String(),
						NumOfConfirmations: decimal.Zero.String(), // Gate.io does not provide this info
						WithdrawPrecision:  withdrawalPrecision.String(),
						Fee:                fee.String(),
						WithdrawFeeType:    models.WithdrawalFeeTypeFixed,
					})
				}
			}
		}
	}
	return withdrawalRules, nil
}

func (o *Service) GetWithdrawalByID(ctx context.Context, args *models.GetWithdrawalByIDParams) (*models.WithdrawalStatusDTO, error) {
	req := &gateio.GetWithdrawalHistoryRequest{Limit: "1"}
	if args.ClientOrderID == nil && args.ExternalOrderID == nil {
		return nil, fmt.Errorf("either ClientOrderID or ExternalOrderID must be provided")
	}
	if args.ClientOrderID != nil {
		if *args.ClientOrderID != "" {
			req.WithdrawOrderID = *args.ClientOrderID
		}
	}
	if args.ExternalOrderID != nil {
		if *args.ExternalOrderID != "" {
			req.WithdrawalID = *args.ExternalOrderID
		}
	}

	res, err := o.exClient.Wallet().GetWithdrawalHistory(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(res.Data) == 0 {
		return nil, fmt.Errorf("withdrawal not found")
	}

	for _, withdrawal := range res.Data {
		if args.ExternalOrderID != nil {
			if withdrawal.ID == *args.ExternalOrderID {
				return &models.WithdrawalStatusDTO{
					ID:           withdrawal.ID,
					TxHash:       withdrawal.Txid,
					NativeAmount: withdrawal.Amount,
					Status:       withdrawal.Status.String(),
				}, nil
			}
		}
		if args.ClientOrderID != nil {
			if withdrawal.ID == *args.ClientOrderID {
				return &models.WithdrawalStatusDTO{
					ID:           withdrawal.ID,
					TxHash:       withdrawal.Txid,
					NativeAmount: withdrawal.Amount,
					Status:       withdrawal.Status.String(),
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("withdrawal ID %s not found", *args.ClientOrderID)
}

func (o *Service) GetConnectionHash() string {
	return o.connHash
}

func (o *Service) GetWithdrawalAddresses(_ context.Context, _ string) ([]*models.WithdrawalAddressDTO, error) {
	return nil, nil
}

func (o *Service) GetSymbolsByCurrency(_ context.Context, _ string) ([]*models.ExchangeSymbolDTO, error) {
	return nil, nil
}

func (o *Service) GetWithdrawalHistory(_ context.Context) error {
	return nil
}

func (o *Service) CreateSpotOrder(ctx context.Context, _ string, _ string, side string, ticker string, _ *decimal.Decimal, rule *models.OrderRulesDTO) (*models.ExchangeOrderDTO, error) {
	tradingSymbol, err := o.exClient.Spot().GetSpotSupportedCurrencyPair(ctx, ticker)
	if err != nil {
		return nil, fmt.Errorf("get trading symbol %s: %w", ticker, err)
	}
	if tradingSymbol.Data.TradeStatus != gateio.PairTradeStatusTradable.String() {
		return nil, fmt.Errorf("trading is disabled for symbol %s", ticker)
	}

	spotOrderRequest := gateio.CreateSpotOrderRequest{
		CurrencyPair: ticker,
		Type:         gateio.OrderTypeMarket.String(),
		Side:         gateio.OrderSide(side).String(),
		TimeInForce:  "fok", // Fill or Kill
	}

	baseBalance := decimal.Zero
	{
		res, err := o.exClient.Spot().GetSpotAccountBalances(ctx, &gateio.GetSpotAccountBalancesRequest{
			Currency: tradingSymbol.Data.Base,
		})
		if err != nil {
			return nil, fmt.Errorf("get base currency balance %s: %w", tradingSymbol.Data.Base, err)
		}
		if len(res.Data) == 0 {
			return nil, fmt.Errorf("base currency %s not found", tradingSymbol.Data.Base)
		}
		baseBalance = baseBalance.Add(res.Data[0].Available)
	}

	quoteBalance := decimal.Zero
	{
		res, err := o.exClient.Spot().GetSpotAccountBalances(ctx, &gateio.GetSpotAccountBalancesRequest{
			Currency: tradingSymbol.Data.Quote,
		})
		if err != nil {
			return nil, fmt.Errorf("get quote currency balance %s: %w", tradingSymbol.Data.Quote, err)
		}
		if len(res.Data) == 0 {
			return nil, fmt.Errorf("quote currency %s not found", tradingSymbol.Data.Quote)
		}
		quoteBalance = quoteBalance.Add(res.Data[0].Available)
	}

	orderMinimumBase, err := decimal.NewFromString(rule.MinOrderAmount)
	if err != nil {
		return nil, err
	}

	orderMinimumQuote, err := decimal.NewFromString(rule.MinOrderValue)
	if err != nil {
		return nil, err
	}

	amount := decimal.Zero
	switch spotOrderRequest.Side {
	case gateio.OrderSideSell.String():
		amount = amount.Add(baseBalance)
		if amount.LessThan(orderMinimumBase) {
			return nil, ErrInsufficientBalance
		}
	case gateio.OrderSideBuy.String():
		amount = amount.Add(quoteBalance)
		if amount.LessThan(orderMinimumQuote) {
			return nil, ErrInsufficientBalance
		}
	default:
		return nil, fmt.Errorf("unsupported order type %s", spotOrderRequest.Side)
	}

	switch spotOrderRequest.Side {
	case gateio.OrderSideSell.String():
		spotOrderRequest.Amount = amount.RoundDown(int32(rule.AmountPrecision)).String() //nolint:gosec
	case gateio.OrderSideBuy.String():
		spotOrderRequest.Amount = amount.RoundDown(int32(rule.ValuePrecision)).String() //nolint:gosec
	}

	placedOrder, err := o.exClient.Spot().CreateSpotOrder(ctx, &spotOrderRequest)
	if err != nil {
		return nil, fmt.Errorf("create spot order: %w", err)
	}

	if placedOrder.Data == nil || placedOrder.Data.ID == "" {
		return nil, fmt.Errorf("failed to create spot order for %s", ticker)
	}

	return &models.ExchangeOrderDTO{
		ExchangeOrderID: placedOrder.Data.ID,
		Amount:          amount,
	}, nil
}

func (o *Service) CreateWithdrawalOrder(ctx context.Context, args *models.CreateWithdrawalOrderParams) (*models.ExchangeWithdrawalDTO, error) {
	args.NativeAmount = args.NativeAmount.RoundDown(int32(args.WithdrawalPrecision)) //nolint:gosec

	internalCurrency, err := o.storage.ExchangeChains().GetTickerByCurrencyID(ctx, repo_exchange_chains.GetTickerByCurrencyIDParams{
		CurrencyID: args.Currency,
		Slug:       models.ExchangeSlugGateio,
	})
	if err != nil {
		return nil, err
	}

	withdrawalWhitelist, err := o.exClient.Wallet().GetWalletSavedAddresses(ctx, &gateio.GetWalletSavedAddressesRequest{
		Currency: internalCurrency,
		Chain:    args.Chain,
	})
	if err != nil {
		return nil, fmt.Errorf("get withdrawal whitelist for %s: %w", args.Currency, err)
	}

	if len(withdrawalWhitelist.Data) == 0 {
		return nil, fmt.Errorf("no withdrawal whitelist found for %s: %w", args.Currency, exchangeclient.ErrWithdrawalAddressNotWhitelisted)
	}

	if !lo.ContainsBy(withdrawalWhitelist.Data, func(item *gateio.SavedAddress) bool {
		return strings.EqualFold(item.Address, args.Address) && strings.EqualFold(item.Chain, args.Chain) && item.Verified.IsVerified()
	}) {
		return nil, fmt.Errorf("withdrawal address %s on chain %s not whitelisted for %s: %w", args.Address, args.Chain, args.Currency, exchangeclient.ErrWithdrawalAddressNotWhitelisted)
	}

	req := &gateio.CreateWithdrawalRequest{
		Currency: internalCurrency,
		Chain:    args.Chain,
		Address:  args.Address,
		Amount:   args.NativeAmount.Sub(args.Fee).String(),
	}

	o.l.Info("withdrawal request assembled",
		"exchange", models.ExchangeSlugGateio.String(),
		"recordID", args.RecordID.String(),
		"request", req,
		"amount", req.Amount,
		"currency", req.Currency,
		"chain", req.Chain,
		"address", req.Address,
	)

	amount := args.NativeAmount.Sub(args.Fee)
	minWithdrawal := args.MinWithdrawal

	dto := &models.ExchangeWithdrawalDTO{}

	for {
		if amount.LessThan(minWithdrawal) {
			o.l.Info("withdrawal amount below minimum",
				"exchange", models.ExchangeSlugOkx.String(),
				"recordID", args.RecordID.String(),
				"current_amount", amount.String(),
				"min_withdrawal", minWithdrawal.String(),
			)
			return nil, exchangeclient.ErrWithdrawalBalanceLocked
		}

		withdrawalStep, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
			Source:     models.ExchangeSlugGateio.String(),
			From:       models.CurrencyCodeUSDT,
			To:         req.Currency,
			Amount:     decimal.NewFromInt(WithdrawalStep).String(),
			StableCoin: false,
		})
		if err != nil {
			return nil, err
		}

		req.Amount = amount.String()
		res, err := o.exClient.Wallet().CreateWithdrawal(ctx, req)
		if err == nil {
			if res.Data.WithdrawOrderID != "" {
				dto.InternalOrderID = res.Data.WithdrawOrderID
			}
			dto.ExternalOrderID = res.Data.ID
			return dto, nil
		}

		if errors.Is(err, exchangeclient.ErrWithdrawalBalanceLocked) {
			o.l.Error("insufficient funds, retrying with reduced amount",
				exchangeclient.ErrWithdrawalBalanceLocked,
				"exchange", models.ExchangeSlugGateio.String(),
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

func (o *Service) GetOrderDetails(ctx context.Context, args *models.GetOrderByIDParams) (*models.OrderDetailsDTO, error) {
	order := &models.OrderDetailsDTO{
		State:     models.ExchangeOrderStatusFailed,
		Amount:    decimal.Zero,
		AmountUSD: decimal.Zero,
	}
	if args.ExternalOrderID != nil { //nolint:nestif
		res, err := o.exClient.Spot().GetSpotOrder(ctx, *args.ExternalOrderID)
		if err != nil {
			return nil, fmt.Errorf("get spot order %s: %w", *args.ExternalOrderID, err)
		}
		if res.Data == nil {
			return order, nil // No order found just leave it as failed
		}

		switch res.Data.Status {
		case gateio.OrderStatusClosed:
			order.State = models.ExchangeOrderStatusCompleted
		case gateio.OrderStatusOpen:
			order.State = models.ExchangeOrderStatusInProgress
		case gateio.OrderStatusCanceled:
			order.State = models.ExchangeOrderStatusFailed
		default:
			order.State = models.ExchangeOrderStatusInProgress
		}

		if res.Data.FinishAs.IsFailed() {
			order.State = models.ExchangeOrderStatusFailed
		}

		pair, err := o.exClient.Spot().GetSpotSupportedCurrencyPair(ctx, res.Data.CurrencyPair)
		if err != nil {
			return nil, fmt.Errorf("get trading symbol %s: %w", res.Data.CurrencyPair, err)
		}

		ticker, err := o.exClient.Spot().GetTickersInfo(ctx, &gateio.GetTickersInfoRequest{
			CurrencyPair: res.Data.CurrencyPair,
		})
		if err != nil {
			return nil, fmt.Errorf("get ticker info for %s: %w", res.Data.CurrencyPair, err)
		}

		orderAmount, err := decimal.NewFromString(res.Data.Amount)
		if err != nil {
			return nil, fmt.Errorf("parse order amount %s: %w", res.Data.Amount, err)
		}
		order.Amount = orderAmount

		switch res.Data.Side {
		case gateio.OrderSideBuy:
			// For market BUY: orderAmount is in QUOTE currency
			// Calculate amount of BASE received = orderAmount / price
			order.Amount = orderAmount.Div(ticker.Data[0].Last).Round(4)

			// Calculate USDT/USDC value
			if pair.Data.Quote == models.CurrencyCodeUSDT || pair.Data.Quote == models.CurrencyCodeUSDC {
				// Quote is already USDT/USDC, so orderAmount is the USDT/USDC value
				order.AmountUSD = orderAmount
				return order, nil
			}

			// If base is USDT/USDC, then order.Amount is already in USD
			if pair.Data.Base == models.CurrencyCodeUSDT || pair.Data.Base == models.CurrencyCodeUSDC {
				order.AmountUSD = order.Amount
				return order, nil
			}

			// Neither base nor quote is USD, need to convert base to USD
			baseUSDPair, err := o.exClient.Spot().GetSpotSupportedCurrencyPair(ctx, pair.Data.Base+"_"+models.CurrencyCodeUSDT)
			if err != nil {
				// Try USDC if USDT fails
				baseUSDPair, err = o.exClient.Spot().GetSpotSupportedCurrencyPair(ctx, pair.Data.Base+"_"+models.CurrencyCodeUSDC)
				if err != nil {
					return nil, fmt.Errorf("get USD trading pair for %s: %w", pair.Data.Base, err)
				}
			}

			baseUSDTicker, err := o.exClient.Spot().GetTickersInfo(ctx, &gateio.GetTickersInfoRequest{
				CurrencyPair: baseUSDPair.Data.Base + "_" + baseUSDPair.Data.Quote,
			})
			if err != nil {
				return nil, fmt.Errorf("get USD ticker for %s: %w", pair.Data.Base, err)
			}

			if len(baseUSDTicker.Data) > 0 {
				// USD value = amount of BASE * price of BASE in USD
				order.AmountUSD = order.Amount.Mul(baseUSDTicker.Data[0].Last).Round(4)
			}

		case gateio.OrderSideSell:
			// For market SELL: orderAmount is in BASE currency
			order.Amount = orderAmount

			// Calculate USD value
			if pair.Data.Base == models.CurrencyCodeUSDT || pair.Data.Base == models.CurrencyCodeUSDC {
				// Base is already USD
				order.AmountUSD = orderAmount
				return order, nil
			}

			if pair.Data.Quote == models.CurrencyCodeUSDT || pair.Data.Quote == models.CurrencyCodeUSDC {
				// Quote is USD, so USD value = base amount * price
				order.AmountUSD = orderAmount.Mul(ticker.Data[0].Last).Round(4)
				return order, nil
			}

			// Neither base nor quote is USD, need to convert base to USD
			baseUSDPair, err := o.exClient.Spot().GetSpotSupportedCurrencyPair(ctx, pair.Data.Base+"_"+models.CurrencyCodeUSDT)
			if err != nil {
				// Try USDC if USDT fails
				baseUSDPair, err = o.exClient.Spot().GetSpotSupportedCurrencyPair(ctx, pair.Data.Base+"_"+models.CurrencyCodeUSDC)
				if err != nil {
					return nil, fmt.Errorf("get USD trading pair for %s: %w", pair.Data.Base, err)
				}
			}

			baseUSDTicker, err := o.exClient.Spot().GetTickersInfo(ctx, &gateio.GetTickersInfoRequest{
				CurrencyPair: baseUSDPair.Data.Base + "_" + baseUSDPair.Data.Quote,
			})
			if err != nil {
				return nil, fmt.Errorf("get USD ticker for %s: %w", pair.Data.Base, err)
			}

			if len(baseUSDTicker.Data) > 0 {
				// USD value = amount of BASE * price of BASE in USD
				order.AmountUSD = orderAmount.Mul(baseUSDTicker.Data[0].Last).Round(4)
			}
		}

		return order, nil
	}
	o.l.Error("failed to get order details", fmt.Errorf("ExternalOrderID is nil"))
	return order, nil
}
