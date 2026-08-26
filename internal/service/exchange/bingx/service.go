package bingx

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
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
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bingx"
	bingxrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/bingx/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bingx/responses"
	"github.com/dv-net/dv-merchant/pkg/iso"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

const (
	symbolSeparator = "-"
	coinCacheTTL    = time.Minute
	WithdrawalStep  = 10
)

type Service struct {
	exClient  *bingx.BaseClient
	storage   storage.IStorage
	convSvc   currconv.ICurrencyConvertor
	l         logger.Logger
	connHash  string
	cacheMu   sync.Mutex
	addresses map[string][]responses.DepositAddress
}

func NewService(logger logger.Logger, apiKey, secretKey string, baseURL *url.URL, storage storage.IStorage, store limiter.Store, convSvc currconv.ICurrencyConvertor) (*Service, error) {
	exClient, err := bingx.NewBaseClient(&bingx.ClientOptions{
		APIKey:    apiKey,
		SecretKey: secretKey,
		BaseURL:   baseURL,
	}, store, bingx.WithLogger(logger))
	if err != nil {
		return nil, err
	}

	connHash, err := hash.SHA256ConnectionHash(models.ExchangeSlugBingx.String(), apiKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection hash: %w", err)
	}

	return &Service{
		exClient: exClient,
		storage:  storage,
		convSvc:  convSvc,
		l:        logger,
		connHash: connHash,
	}, nil
}

func (o *Service) GetConnectionHash() string { return o.connHash }

func (o *Service) TestConnection(ctx context.Context) error {
	if _, err := o.exClient.Account().GetBalance(ctx); err != nil {
		return fmt.Errorf("get account balance: %w", err)
	}
	return nil
}

var (
	coinsCacheMu sync.Mutex
	coinsCache   = make(map[string]cachedCoins)
)

type cachedCoins struct {
	value responses.GetCoinsConfigResponse
	at    time.Time
}

func (o *Service) coinsConfig(ctx context.Context) (responses.GetCoinsConfigResponse, error) {
	coinsCacheMu.Lock()
	defer coinsCacheMu.Unlock()

	if cached, ok := coinsCache[o.connHash]; ok && time.Since(cached.at) < coinCacheTTL {
		return cached.value, nil
	}

	coins, err := o.exClient.Wallet().GetCoinsConfig(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("get coins config: %w", err)
	}
	coinsCache[o.connHash] = cachedCoins{value: coins, at: time.Now()}

	return coins, nil
}

func (o *Service) coinDepositAddresses(ctx context.Context, coin string) ([]responses.DepositAddress, error) {
	o.cacheMu.Lock()
	defer o.cacheMu.Unlock()

	if cached, ok := o.addresses[coin]; ok {
		return cached, nil
	}

	res, err := o.exClient.Wallet().GetDepositAddress(ctx, &bingxrequests.GetDepositAddressRequest{Coin: coin})
	if err != nil {
		return nil, fmt.Errorf("get deposit addresses for %s: %w", coin, err)
	}

	if o.addresses == nil {
		o.addresses = make(map[string][]responses.DepositAddress)
	}
	o.addresses[coin] = res.Data

	return res.Data, nil
}

func (o *Service) GetAccountBalance(ctx context.Context) ([]*models.AccountBalanceDTO, error) {
	account, err := o.exClient.Account().GetBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("get account balance: %w", err)
	}

	fund, err := o.exClient.Account().GetFundBalance(ctx)
	if err != nil {
		return nil, fmt.Errorf("get fund balance: %w", err)
	}

	totals := make(map[string]decimal.Decimal, len(account.Balances)+len(fund.Assets))
	order := make([]string, 0, len(account.Balances)+len(fund.Assets))
	for _, balance := range account.Balances {
		if _, exists := totals[balance.Asset]; !exists {
			order = append(order, balance.Asset)
		}
		totals[balance.Asset] = totals[balance.Asset].Add(balance.Free)
	}
	for _, asset := range fund.Assets {
		if _, exists := totals[asset.Asset]; !exists {
			order = append(order, asset.Asset)
		}
		totals[asset.Asset] = totals[asset.Asset].Add(asset.Free)
	}

	balances := make([]*models.AccountBalanceDTO, 0, len(order))
	for _, ticker := range order {
		amount := totals[ticker]
		if amount.IsZero() {
			continue
		}

		currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByTicker(ctx, ticker)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, fmt.Errorf("get internal currency id for %s: %w", ticker, err)
		}

		amountUSD, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
			Source: models.ExchangeSlugBingx.String(),
			From:   ticker,
			To:     models.CurrencyCodeUSDT,
			Amount: amount.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("convert %s balance to usdt: %w", ticker, err)
		}

		balances = append(balances, &models.AccountBalanceDTO{
			Currency:  currencyID,
			Amount:    amount,
			AmountUSD: amountUSD.Round(4),
			Type:      models.CurrencyTypeCrypto.String(),
		})
	}

	return balances, nil
}

func (o *Service) spotBalance(ctx context.Context, currency string) (decimal.Decimal, error) {
	account, err := o.exClient.Account().GetBalance(ctx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("get account balance: %w", err)
	}

	for _, balance := range account.Balances {
		if balance.Asset == currency {
			return balance.Free, nil
		}
	}

	return decimal.Zero, nil
}

func (o *Service) GetCurrencyBalance(ctx context.Context, currency string) (*decimal.Decimal, error) {
	spot, err := o.spotBalance(ctx, currency)
	if err != nil {
		return nil, err
	}

	fund, err := o.fundBalance(ctx, currency)
	if err != nil {
		return nil, err
	}

	total := spot.Add(fund)

	return &total, nil
}

func isTradable(symbol responses.SymbolInfo, side string) bool {
	if symbol.IsDelisted() {
		return false
	}
	switch strings.ToUpper(side) {
	case bingx.OrderSideBuy:
		return symbol.APIStateBuy
	case bingx.OrderSideSell:
		return symbol.APIStateSell
	default:
		return symbol.APIStateBuy && symbol.APIStateSell
	}
}

func splitSymbol(symbol string) (base, quote string, ok bool) {
	parts := strings.SplitN(symbol, symbolSeparator, 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func precisionOf(step decimal.Decimal) int {
	if step.IsZero() {
		return 0
	}
	if exp := step.Exponent(); exp < 0 {
		return int(-exp)
	}
	return 0
}

func roundToStep(amount, step decimal.Decimal) decimal.Decimal {
	if !step.IsPositive() {
		return amount
	}
	return amount.Div(step).Floor().Mul(step)
}

func (o *Service) fundBalance(ctx context.Context, currency string) (decimal.Decimal, error) {
	fund, err := o.exClient.Account().GetFundBalance(ctx)
	if err != nil {
		return decimal.Zero, fmt.Errorf("get fund balance: %w", err)
	}

	for _, asset := range fund.Assets {
		if asset.Asset == currency {
			return asset.Free, nil
		}
	}

	return decimal.Zero, nil
}

func (o *Service) ensureSpotBalance(ctx context.Context, currency string, required decimal.Decimal) (decimal.Decimal, error) {
	spot, err := o.spotBalance(ctx, currency)
	if err != nil {
		return decimal.Zero, err
	}

	fund, err := o.fundBalance(ctx, currency)
	if err != nil {
		return decimal.Zero, err
	}

	total := spot.Add(fund)
	if total.LessThan(required) {
		return total, exchangeclient.ErrInsufficientBalance
	}
	if !fund.IsPositive() {
		return total, nil
	}

	if _, err := o.exClient.Account().Transfer(ctx, &bingxrequests.TransferRequest{
		Type:   bingx.TransferFundToSpot,
		Asset:  currency,
		Amount: fund.String(),
	}); err != nil {
		return decimal.Zero, fmt.Errorf("transfer %s from fund to spot: %w", currency, err)
	}

	o.l.Infow(
		"moved funds to spot wallet",
		"exchange", models.ExchangeSlugBingx.String(),
		"currency", currency,
		"amount", fund.String(),
	)

	return total, nil
}

func (o *Service) GetExchangeSymbols(ctx context.Context) ([]*models.ExchangeSymbolDTO, error) {
	info, err := o.exClient.Market().GetSymbols(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("get exchange symbols: %w", err)
	}

	symbols := make([]*models.ExchangeSymbolDTO, 0, len(info.Symbols)*2)
	for _, symbol := range info.Symbols {
		base, quote, ok := splitSymbol(symbol.Symbol)
		if !ok || iso.IsFiat(base) || iso.IsFiat(quote) {
			continue
		}

		if isTradable(symbol, bingx.OrderSideSell) {
			symbols = append(symbols, &models.ExchangeSymbolDTO{
				Symbol:      symbol.Symbol,
				DisplayName: base + "/" + quote,
				BaseSymbol:  base,
				QuoteSymbol: quote,
				Type:        models.OrderSideSell.String(),
			})
		}
		if isTradable(symbol, bingx.OrderSideBuy) {
			symbols = append(symbols, &models.ExchangeSymbolDTO{
				Symbol:      symbol.Symbol,
				DisplayName: quote + "/" + base,
				BaseSymbol:  base,
				QuoteSymbol: quote,
				Type:        models.OrderSideBuy.String(),
			})
		}
	}

	return symbols, nil
}

func (o *Service) GetOrderRule(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	info, err := o.exClient.Market().GetSymbols(ctx, ticker)
	if err != nil {
		if errors.Is(err, exchangeclient.ErrRateLimited) {
			return nil, exchangeclient.ErrSkipOrder
		}
		return nil, fmt.Errorf("get order rule for %s: %w", ticker, err)
	}
	if len(info.Symbols) == 0 {
		return nil, fmt.Errorf("symbol %s not found", ticker)
	}

	symbol := info.Symbols[0]
	if symbol.IsDelisted() {
		return nil, exchangeclient.ErrSymbolTradingHalted
	}

	base, quote, ok := splitSymbol(symbol.Symbol)
	if !ok {
		return nil, fmt.Errorf("unexpected symbol format %s", symbol.Symbol)
	}

	return &models.OrderRulesDTO{
		Symbol:                 symbol.Symbol,
		BaseCurrency:           base,
		QuoteCurrency:          quote,
		MinOrderAmount:         symbol.MinQty.String(),
		MaxOrderAmount:         symbol.MaxQty.String(),
		MinOrderValue:          symbol.MinNotional.String(),
		BuyMarketMaxOrderValue: symbol.MaxMarketNotional.String(),
		AmountPrecision:        precisionOf(symbol.StepSize),
		PricePrecision:         precisionOf(symbol.TickSize),
		ValuePrecision:         precisionOf(symbol.TickSize),
	}, nil
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

func (o *Service) CreateSpotOrder(ctx context.Context, _ string, _ string, side string, ticker string, _ *decimal.Decimal, rule *models.OrderRulesDTO) (*models.ExchangeOrderDTO, error) {
	req := &bingxrequests.PlaceOrderRequest{
		Symbol: ticker,
		Side:   strings.ToUpper(side),
		Type:   bingx.OrderTypeMarket,
	}

	info, err := o.exClient.Market().GetSymbols(ctx, ticker)
	if err != nil {
		if errors.Is(err, exchangeclient.ErrRateLimited) {
			return nil, exchangeclient.ErrSkipOrder
		}
		return nil, fmt.Errorf("get symbol info for %s: %w", ticker, err)
	}
	if len(info.Symbols) == 0 {
		return nil, fmt.Errorf("symbol %s not found", ticker)
	}
	if !isTradable(info.Symbols[0], req.Side) {
		return nil, exchangeclient.ErrSymbolTradingHalted
	}

	clientOrderID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	req.ClientOrderID = strings.ReplaceAll(clientOrderID.String(), "-", "")

	maxMarketValue, err := decimal.NewFromString(rule.BuyMarketMaxOrderValue)
	if err != nil {
		return nil, fmt.Errorf("parse max market order value: %w", err)
	}

	var amount decimal.Decimal
	switch req.Side {
	case bingx.OrderSideSell:
		minOrderAmount, err := decimal.NewFromString(rule.MinOrderAmount)
		if err != nil {
			return nil, fmt.Errorf("parse min order amount: %w", err)
		}
		balance, err := o.ensureSpotBalance(ctx, rule.BaseCurrency, minOrderAmount)
		if err != nil {
			if errors.Is(err, exchangeclient.ErrRateLimited) {
				return nil, exchangeclient.ErrSkipOrder
			}
			return nil, err
		}
		amount = roundToStep(balance, info.Symbols[0].StepSize)

		ticker, err := o.exClient.Market().GetTicker(ctx, req.Symbol)
		if err != nil {
			if errors.Is(err, exchangeclient.ErrRateLimited) {
				return nil, exchangeclient.ErrSkipOrder
			}
			return nil, fmt.Errorf("get ticker price for %s: %w", req.Symbol, err)
		}
		if len(ticker) == 0 || !ticker[0].LastPrice.IsPositive() {
			return nil, exchangeclient.ErrSkipOrder
		}

		minOrderValue, err := decimal.NewFromString(rule.MinOrderValue)
		if err != nil {
			return nil, fmt.Errorf("parse min order value: %w", err)
		}
		orderValue := amount.Mul(ticker[0].LastPrice)
		if orderValue.LessThan(minOrderValue) {
			return nil, exchangeclient.ErrInsufficientBalance
		}
		if maxMarketValue.IsPositive() && orderValue.GreaterThan(maxMarketValue) {
			amount = roundToStep(maxMarketValue.Div(ticker[0].LastPrice), info.Symbols[0].StepSize)
		}

		req.Quantity = amount.String()
	case bingx.OrderSideBuy:
		minOrderValue, err := decimal.NewFromString(rule.MinOrderValue)
		if err != nil {
			return nil, fmt.Errorf("parse min order value: %w", err)
		}
		balance, err := o.ensureSpotBalance(ctx, rule.QuoteCurrency, minOrderValue)
		if err != nil {
			if errors.Is(err, exchangeclient.ErrRateLimited) {
				return nil, exchangeclient.ErrSkipOrder
			}
			return nil, err
		}
		amount = balance.RoundDown(int32(rule.ValuePrecision)) //nolint:gosec
		if maxMarketValue.IsPositive() && amount.GreaterThan(maxMarketValue) {
			amount = maxMarketValue.RoundDown(int32(rule.ValuePrecision)) //nolint:gosec
		}
		req.QuoteOrderQty = amount.String()
	default:
		return nil, fmt.Errorf("unsupported order side %s", req.Side)
	}

	order, err := o.exClient.Spot().PlaceOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("place spot order: %w", err)
	}
	if order.OrderID == 0 {
		return nil, fmt.Errorf("failed to create spot order for %s", ticker)
	}

	return &models.ExchangeOrderDTO{
		ExchangeOrderID: strconv.FormatInt(order.OrderID, 10),
		ClientOrderID:   req.ClientOrderID,
		Amount:          amount,
	}, nil
}

func (o *Service) GetOrderDetails(ctx context.Context, args *models.GetOrderByIDParams) (*models.OrderDetailsDTO, error) {
	order := &models.OrderDetailsDTO{
		State:     models.ExchangeOrderStatusFailed,
		Amount:    decimal.Zero,
		AmountUSD: decimal.Zero,
	}

	if args.InstrumentID == nil || *args.InstrumentID == "" {
		return order, fmt.Errorf("instrument id is required")
	}

	req := &bingxrequests.QueryOrderRequest{Symbol: *args.InstrumentID}
	switch {
	case args.ExternalOrderID != nil && *args.ExternalOrderID != "":
		req.OrderID = *args.ExternalOrderID
	case args.ClientOrderID != nil && *args.ClientOrderID != "":
		req.ClientOrderID = *args.ClientOrderID
	default:
		return order, fmt.Errorf("either external or client order id must be provided")
	}

	res, err := o.exClient.Spot().QueryOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("query order: %w", err)
	}

	switch res.Status {
	case responses.OrderStatusFilled:
		order.State = models.ExchangeOrderStatusCompleted
	case responses.OrderStatusNew, responses.OrderStatusPending, responses.OrderStatusPartiallyFilled:
		order.State = models.ExchangeOrderStatusInProgress
	case responses.OrderStatusCanceled, responses.OrderStatusFailed:
		order.State = models.ExchangeOrderStatusFailed
	default:
		order.State = models.ExchangeOrderStatusInProgress
	}

	order.Amount = res.ExecutedQty

	_, quote, ok := splitSymbol(res.Symbol)
	if !ok {
		return nil, fmt.Errorf("unexpected symbol format %s", res.Symbol)
	}

	if quote == models.CurrencyCodeUSDT || quote == models.CurrencyCodeUSDC {
		order.AmountUSD = res.CummulativeQuoteQty
		return order, nil
	}

	amountUSD, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
		Source: models.ExchangeSlugBingx.String(),
		From:   quote,
		To:     models.CurrencyCodeUSDT,
		Amount: res.CummulativeQuoteQty.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("convert %s to usdt: %w", quote, err)
	}
	order.AmountUSD = amountUSD.Round(4)

	return order, nil
}

func (o *Service) GetDepositAddresses(ctx context.Context, currency, chain string) ([]*models.DepositAddressDTO, error) {
	addresses, err := o.coinDepositAddresses(ctx, currency)
	if err != nil {
		if errors.Is(err, bingx.ErrDepositAddressUnavailable) {
			return nil, nil
		}
		return nil, err
	}

	matched := filterByNetwork(addresses, chain)
	if len(matched) == 0 {
		return nil, nil
	}

	currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByParams(ctx, repo_exchange_chains.GetCurrencyIDByParamsParams{
		Ticker: currency,
		Chain:  chain,
		Slug:   models.ExchangeSlugBingx,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get internal currency id for %s: %w", chain, err)
	}

	return lo.Map(matched, func(item responses.DepositAddress, _ int) *models.DepositAddressDTO {
		return &models.DepositAddressDTO{
			Address:          item.Address,
			Currency:         currencyID,
			Chain:            chain,
			InternalCurrency: item.Coin,
			AddressType:      models.DepositAddress,
			PaymentTag:       item.Tag,
		}
	}), nil
}

func filterByNetwork(addresses []responses.DepositAddress, chain string) []responses.DepositAddress {
	return lo.Filter(addresses, func(item responses.DepositAddress, _ int) bool {
		return item.Network == chain
	})
}

func (o *Service) GetWithdrawalRules(ctx context.Context, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	currEnabled, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugBingx)
	if err != nil {
		return nil, fmt.Errorf("get enabled currencies: %w", err)
	}

	currEnabled = lo.Filter(currEnabled, func(item *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) bool {
		return lo.Contains(currencies, item.ID.String)
	})

	coins, err := o.coinsConfig(ctx)
	if err != nil {
		return nil, err
	}

	rules := make([]*models.WithdrawalRulesDTO, 0, len(currEnabled))
	for _, coin := range coins {
		for _, network := range coin.NetworkList {
			if !network.WithdrawEnable {
				continue
			}
			if !lo.ContainsBy(currEnabled, func(item *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
				return item.Ticker == coin.Coin && item.Chain == network.Network
			}) {
				continue
			}

			minDepositAmount := network.DepositMin
			if minDepositAmount.IsZero() {
				converted, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
					Source: models.ExchangeSlugBingx.String(),
					From:   models.CurrencyCodeUSDT,
					To:     coin.Coin,
					Amount: "1",
				})
				if err != nil {
					return nil, fmt.Errorf("convert minimum deposit amount for %s: %w", coin.Coin, err)
				}
				minDepositAmount = converted
			}

			rules = append(rules, &models.WithdrawalRulesDTO{
				Currency:           coin.Coin,
				Chain:              network.Network,
				MinDepositAmount:   minDepositAmount.String(),
				MinWithdrawAmount:  network.WithdrawMin.String(),
				MaxWithdrawAmount:  network.WithdrawMax.String(),
				NumOfConfirmations: strconv.FormatInt(network.MinConfirm, 10),
				WithdrawPrecision:  strconv.FormatInt(int64(network.WithdrawPrecision), 10),
				WithdrawFeeType:    models.WithdrawalFeeTypeFixed,
				Fee:                network.WithdrawFee.String(),
			})
		}
	}

	return rules, nil
}

func (o *Service) ensureFundBalance(ctx context.Context, currency string, required decimal.Decimal) (decimal.Decimal, error) {
	fund, err := o.fundBalance(ctx, currency)
	if err != nil {
		return decimal.Zero, err
	}

	spot, err := o.spotBalance(ctx, currency)
	if err != nil {
		return decimal.Zero, err
	}

	total := fund.Add(spot)
	if total.LessThan(required) {
		return total, exchangeclient.ErrInsufficientBalance
	}
	if !spot.IsPositive() {
		return total, nil
	}

	if _, err := o.exClient.Account().Transfer(ctx, &bingxrequests.TransferRequest{
		Type:   bingx.TransferSpotToFund,
		Asset:  currency,
		Amount: spot.String(),
	}); err != nil {
		return decimal.Zero, fmt.Errorf("transfer %s from spot to fund: %w", currency, err)
	}

	o.l.Infow(
		"moved funds to fund wallet",
		"exchange", models.ExchangeSlugBingx.String(),
		"currency", currency,
		"amount", spot.String(),
	)

	return total, nil
}

func (o *Service) CreateWithdrawalOrder(ctx context.Context, args *models.CreateWithdrawalOrderParams) (*models.ExchangeWithdrawalDTO, error) {
	args.NativeAmount = args.NativeAmount.RoundDown(int32(args.WithdrawalPrecision)) //nolint:gosec

	internalCurrency, err := o.storage.ExchangeChains().GetTickerByCurrencyID(ctx, repo_exchange_chains.GetTickerByCurrencyIDParams{
		CurrencyID: args.Currency,
		Slug:       models.ExchangeSlugBingx,
	})
	if err != nil {
		return nil, fmt.Errorf("get exchange ticker for %s: %w", args.Currency, err)
	}

	if args.NativeAmount.LessThan(args.MinWithdrawal) {
		return nil, exchangeclient.ErrMinWithdrawalBalance
	}

	amount := args.NativeAmount.Sub(args.Fee).RoundDown(int32(args.WithdrawalPrecision)) //nolint:gosec
	if !amount.IsPositive() {
		return nil, exchangeclient.ErrMinWithdrawalBalance
	}
	minWithdrawal := args.MinWithdrawal

	if _, err := o.ensureFundBalance(ctx, internalCurrency, args.NativeAmount); err != nil {
		return nil, err
	}

	o.l.Infow(
		"withdrawal request assembled",
		"exchange", models.ExchangeSlugBingx.String(),
		"recordID", args.RecordID.String(),
		"amount", amount.String(),
		"fee", args.Fee.String(),
		"currency", internalCurrency,
		"chain", args.Chain,
		"address", args.Address,
	)

	dto := &models.ExchangeWithdrawalDTO{}

	for {
		withdrawalStep, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
			Source:     models.ExchangeSlugBingx.String(),
			From:       models.CurrencyCodeUSDT,
			To:         internalCurrency,
			Amount:     decimal.NewFromInt(WithdrawalStep).String(),
			StableCoin: false,
		})
		if err != nil {
			return nil, err
		}

		order, err := o.exClient.Wallet().Withdraw(ctx, &bingxrequests.WithdrawRequest{
			Coin:       internalCurrency,
			Network:    args.Chain,
			Address:    args.Address,
			Amount:     amount.String(),
			WalletType: bingx.WalletTypeFund,
		})
		if err == nil {
			if order.ID == "" {
				return nil, fmt.Errorf("withdrawal for %s created without id", internalCurrency)
			}

			dto.ExternalOrderID = order.ID
			return dto, nil
		}

		if errors.Is(err, exchangeclient.ErrInsufficientBalance) {
			o.l.Errorw("insufficient funds, retrying with reduced amount",
				"error", exchangeclient.ErrInsufficientBalance,
				"exchange", models.ExchangeSlugBingx.String(),
				"recordID", args.RecordID.String(),
				"current_amount", amount.String(),
			)

			amount = amount.Sub(withdrawalStep).RoundDown(int32(args.WithdrawalPrecision)) //nolint:gosec
			if amount.LessThan(minWithdrawal) {
				return nil, exchangeclient.ErrMinWithdrawalBalance
			}
			dto.RetryReason = exchangeclient.ErrInsufficientBalance.Error()
			continue
		}

		return nil, err
	}
}

func (o *Service) GetWithdrawalByID(ctx context.Context, args *models.GetWithdrawalByIDParams) (*models.WithdrawalStatusDTO, error) {
	orderID := ""
	switch {
	case args.ExternalOrderID != nil && *args.ExternalOrderID != "":
		orderID = *args.ExternalOrderID
	case args.ClientOrderID != nil && *args.ClientOrderID != "":
		orderID = *args.ClientOrderID
	default:
		return nil, fmt.Errorf("either ClientOrderID or ExternalOrderID must be provided")
	}

	history, err := o.exClient.Wallet().GetWithdrawHistory(ctx, &bingxrequests.GetWithdrawHistoryRequest{})
	if err != nil {
		return nil, fmt.Errorf("get withdrawal history: %w", err)
	}

	for _, withdrawal := range history {
		if withdrawal.ID != orderID {
			continue
		}
		return &models.WithdrawalStatusDTO{
			ID:           withdrawal.ID,
			TxHash:       withdrawal.TxID,
			NativeAmount: withdrawal.Amount,
			Status:       strconv.FormatInt(withdrawal.Status, 10),
		}, nil
	}

	return nil, fmt.Errorf("withdrawal %s not found", orderID)
}
