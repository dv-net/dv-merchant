package mexc

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
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
	mexc "github.com/dv-net/dv-merchant/pkg/exchange_client/mexc"
	mexcrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/mexc/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/mexc/responses"
	"github.com/dv-net/dv-merchant/pkg/iso"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

const coinCacheTTL = time.Minute

type cachedAddresses struct {
	value responses.GetDepositAddressResponse
	at    time.Time
}

type Service struct {
	exClient  *mexc.BaseClient
	storage   storage.IStorage
	convSvc   currconv.ICurrencyConvertor
	l         logger.Logger
	connHash  string
	cacheMu   sync.Mutex
	coins     responses.GetCoinsConfigResponse
	coinsAt   time.Time
	addresses map[string]cachedAddresses
}

func (o *Service) coinDepositAddresses(ctx context.Context, coin string, refresh bool) (responses.GetDepositAddressResponse, error) {
	o.cacheMu.Lock()
	defer o.cacheMu.Unlock()

	if cached, ok := o.addresses[coin]; ok && !refresh && time.Since(cached.at) < coinCacheTTL {
		return cached.value, nil
	}

	addresses, err := o.exClient.Wallet().GetDepositAddress(ctx, &mexcrequests.GetDepositAddressRequest{Coin: coin})
	if err != nil {
		return nil, fmt.Errorf("get deposit addresses for %s: %w", coin, err)
	}

	if o.addresses == nil {
		o.addresses = make(map[string]cachedAddresses)
	}
	o.addresses[coin] = cachedAddresses{value: addresses, at: time.Now()}

	return addresses, nil
}

func (o *Service) coinsConfig(ctx context.Context) (responses.GetCoinsConfigResponse, error) {
	o.cacheMu.Lock()
	defer o.cacheMu.Unlock()

	if o.coins != nil && time.Since(o.coinsAt) < coinCacheTTL {
		return o.coins, nil
	}

	coins, err := o.exClient.Wallet().GetCoinsConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get coins config: %w", err)
	}
	o.coins, o.coinsAt = coins, time.Now()

	return coins, nil
}

func NewService(logger logger.Logger, apiKey, secretKey string, baseURL *url.URL, storage storage.IStorage, store limiter.Store, convSvc currconv.ICurrencyConvertor) (*Service, error) {
	exClient, err := mexc.NewBaseClient(&mexc.ClientOptions{
		APIKey:    apiKey,
		SecretKey: secretKey,
		BaseURL:   baseURL,
	}, store, mexc.WithLogger(logger))
	if err != nil {
		return nil, err
	}

	connHash, err := hash.SHA256ConnectionHash(models.ExchangeSlugMexc.String(), apiKey, secretKey)
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

func (o *Service) TestConnection(ctx context.Context) error {
	if _, err := o.exClient.Account().GetAccountInfo(ctx); err != nil {
		return fmt.Errorf("get account info: %w", err)
	}
	return nil
}

func (o *Service) GetConnectionHash() string {
	return o.connHash
}

func (o *Service) GetAccountBalance(ctx context.Context) ([]*models.AccountBalanceDTO, error) {
	account, err := o.exClient.Account().GetAccountInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("get account info: %w", err)
	}

	balances := make([]*models.AccountBalanceDTO, 0, len(account.Balances))
	for _, balance := range account.Balances {
		if balance.Free.IsZero() {
			continue
		}

		currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByTicker(ctx, balance.Asset)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return nil, fmt.Errorf("get internal currency id for %s: %w", balance.Asset, err)
		}

		amountUSD, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
			Source: models.ExchangeSlugMexc.String(),
			From:   balance.Asset,
			To:     models.CurrencyCodeUSDT,
			Amount: balance.Free.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("convert %s balance to usdt: %w", balance.Asset, err)
		}

		balances = append(balances, &models.AccountBalanceDTO{
			Currency:  currencyID,
			Amount:    balance.Free,
			AmountUSD: amountUSD.Round(4),
			Type:      models.CurrencyTypeCrypto.String(),
		})
	}

	return balances, nil
}

func (o *Service) GetCurrencyBalance(ctx context.Context, currency string) (*decimal.Decimal, error) {
	account, err := o.exClient.Account().GetAccountInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("get account info: %w", err)
	}

	for _, balance := range account.Balances {
		if balance.Asset == currency {
			return &balance.Free, nil
		}
	}

	return &decimal.Zero, nil
}

func (o *Service) GetExchangeSymbols(ctx context.Context) ([]*models.ExchangeSymbolDTO, error) {
	info, err := o.exClient.Market().GetExchangeInfo(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("get exchange info: %w", err)
	}

	symbols := make([]*models.ExchangeSymbolDTO, 0, len(info.Symbols)*2)
	for _, symbol := range info.Symbols {
		if !isTradable(symbol) {
			continue
		}
		if iso.IsFiat(symbol.BaseAsset) || iso.IsFiat(symbol.QuoteAsset) {
			continue
		}

		symbols = append(symbols, &models.ExchangeSymbolDTO{
			Symbol:      symbol.Symbol,
			DisplayName: symbol.BaseAsset + "/" + symbol.QuoteAsset,
			BaseSymbol:  symbol.BaseAsset,
			QuoteSymbol: symbol.QuoteAsset,
			Type:        models.OrderSideSell.String(),
		}, &models.ExchangeSymbolDTO{
			Symbol:      symbol.Symbol,
			DisplayName: symbol.QuoteAsset + "/" + symbol.BaseAsset,
			BaseSymbol:  symbol.BaseAsset,
			QuoteSymbol: symbol.QuoteAsset,
			Type:        models.OrderSideBuy.String(),
		})
	}

	return symbols, nil
}

func isTradable(symbol responses.SymbolInfo) bool {
	if symbol.Status != mexc.SymbolStatusEnabled || !symbol.IsSpotTradingAllowed {
		return false
	}
	return slices.Contains(symbol.OrderTypes, mexc.OrderTypeMarket)
}

func (o *Service) GetDepositAddresses(ctx context.Context, currency, chain string) ([]*models.DepositAddressDTO, error) {
	addresses, err := o.coinDepositAddresses(ctx, currency, false)
	if err != nil {
		return nil, err
	}

	matched := filterByNetwork(addresses, chain)
	if len(matched) == 0 {
		addresses, err = o.coinDepositAddresses(ctx, currency, true)
		if err != nil {
			return nil, err
		}
		matched = filterByNetwork(addresses, chain)
	}

	if len(matched) == 0 {
		network, err := o.resolveNetworkName(ctx, currency, chain)
		if err != nil {
			return nil, err
		}

		if _, err = o.exClient.Wallet().CreateDepositAddress(ctx, &mexcrequests.CreateDepositAddressRequest{
			Coin:    currency,
			Network: network,
		}); err != nil && !errors.Is(err, mexc.ErrDepositAddressExists) {
			return nil, fmt.Errorf("create deposit address for %s: %w", currency, err)
		}

		addresses, err = o.coinDepositAddresses(ctx, currency, true)
		if err != nil {
			return nil, err
		}

		matched = filterByNetwork(addresses, chain)
		if len(matched) == 0 {
			return nil, nil
		}
	}

	currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByParams(ctx, repo_exchange_chains.GetCurrencyIDByParamsParams{
		Ticker: currency,
		Chain:  chain,
		Slug:   models.ExchangeSlugMexc,
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
			PaymentTag:       item.Memo,
		}
	}), nil
}

func filterByNetwork(addresses responses.GetDepositAddressResponse, chain string) []responses.DepositAddress {
	matched := lo.Filter(addresses, func(item responses.DepositAddress, _ int) bool {
		return item.NetWork == chain
	})

	return lo.UniqBy(matched, func(item responses.DepositAddress) string {
		return fmt.Sprintf("%s:%s", item.Address, item.Memo)
	})
}

func (o *Service) resolveNetworkName(ctx context.Context, currency, chain string) (string, error) {
	coins, err := o.coinsConfig(ctx)
	if err != nil {
		return "", err
	}

	for _, coin := range coins {
		if coin.Coin != currency {
			continue
		}
		for _, network := range coin.NetworkList {
			if network.NetWork == chain {
				return network.Network, nil
			}
		}
	}

	return "", fmt.Errorf("network %s is not supported for %s on mexc", chain, currency)
}

func (o *Service) CreateWithdrawalOrder(ctx context.Context, args *models.CreateWithdrawalOrderParams) (*models.ExchangeWithdrawalDTO, error) {
	args.NativeAmount = args.NativeAmount.RoundDown(int32(args.WithdrawalPrecision)) //nolint:gosec

	internalCurrency, err := o.storage.ExchangeChains().GetTickerByCurrencyID(ctx, repo_exchange_chains.GetTickerByCurrencyIDParams{
		CurrencyID: args.Currency,
		Slug:       models.ExchangeSlugMexc,
	})
	if err != nil {
		return nil, fmt.Errorf("get exchange ticker for %s: %w", args.Currency, err)
	}

	if args.NativeAmount.LessThan(args.MinWithdrawal) {
		return nil, exchangeclient.ErrMinWithdrawalBalance
	}

	o.l.Infow(
		"withdrawal request assembled",
		"exchange", models.ExchangeSlugMexc.String(),
		"recordID", args.RecordID.String(),
		"amount", args.NativeAmount.String(),
		"fee", args.Fee.String(),
		"currency", internalCurrency,
		"chain", args.Chain,
		"address", args.Address,
	)

	order, err := o.exClient.Wallet().Withdraw(ctx, &mexcrequests.WithdrawRequest{
		Coin:    internalCurrency,
		Network: args.Chain,
		Address: args.Address,
		Amount:  args.NativeAmount.String(),
	})
	if err != nil {
		return nil, err
	}
	if order.ID == "" {
		return nil, fmt.Errorf("withdrawal for %s created without id", internalCurrency)
	}

	return &models.ExchangeWithdrawalDTO{ExternalOrderID: order.ID}, nil
}

func (o *Service) CreateSpotOrder(ctx context.Context, _ string, _ string, side string, ticker string, _ *decimal.Decimal, rule *models.OrderRulesDTO) (*models.ExchangeOrderDTO, error) {
	req := &mexcrequests.PlaceOrderRequest{
		Symbol: ticker,
		Side:   strings.ToUpper(side),
		Type:   mexc.OrderTypeMarket,
	}

	info, err := o.exClient.Market().GetExchangeInfo(ctx, ticker)
	if err != nil {
		return nil, fmt.Errorf("get exchange info for %s: %w", ticker, err)
	}
	if len(info.Symbols) == 0 {
		return nil, fmt.Errorf("symbol %s not found", ticker)
	}
	if !isTradable(info.Symbols[0]) || !isSideAllowed(info.Symbols[0], req.Side) {
		return nil, exchangeclient.ErrSymbolTradingHalted
	}

	clientOrderID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	req.NewClientOrderID = strings.ReplaceAll(clientOrderID.String(), "-", "")

	var amount decimal.Decimal
	switch req.Side {
	case mexc.OrderSideSell:
		balance, err := o.GetCurrencyBalance(ctx, rule.BaseCurrency)
		if err != nil {
			return nil, fmt.Errorf("get base currency balance %s: %w", rule.BaseCurrency, err)
		}
		minOrderAmount, err := decimal.NewFromString(rule.MinOrderAmount)
		if err != nil {
			return nil, fmt.Errorf("parse min order amount: %w", err)
		}
		if balance.LessThan(minOrderAmount) {
			return nil, exchangeclient.ErrInsufficientBalance
		}
		amount = roundToSizeStep(*balance, rule.MinOrderAmount, rule.AmountPrecision)

		minOrderValue, err := decimal.NewFromString(rule.MinOrderValue)
		if err != nil {
			return nil, fmt.Errorf("parse min order value: %w", err)
		}
		ticker, err := o.exClient.Market().GetTickerPrice(ctx, req.Symbol)
		if err != nil {
			return nil, fmt.Errorf("get ticker price for %s: %w", req.Symbol, err)
		}
		if !ticker.Price.IsPositive() {
			return nil, exchangeclient.ErrSkipOrder
		}
		if amount.Mul(ticker.Price).LessThan(minOrderValue) {
			return nil, exchangeclient.ErrInsufficientBalance
		}

		req.Quantity = amount.String()
	case mexc.OrderSideBuy:
		balance, err := o.GetCurrencyBalance(ctx, rule.QuoteCurrency)
		if err != nil {
			return nil, fmt.Errorf("get quote currency balance %s: %w", rule.QuoteCurrency, err)
		}
		minOrderValue, err := decimal.NewFromString(rule.MinOrderValue)
		if err != nil {
			return nil, fmt.Errorf("parse min order value: %w", err)
		}
		if balance.LessThan(minOrderValue) {
			return nil, exchangeclient.ErrInsufficientBalance
		}
		amount = balance.RoundDown(int32(rule.ValuePrecision)) //nolint:gosec
		req.QuoteOrderQty = amount.String()
	default:
		return nil, fmt.Errorf("unsupported order side %s", req.Side)
	}

	order, err := o.exClient.Spot().PlaceOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("place spot order: %w", err)
	}
	if order.OrderID == "" {
		return nil, fmt.Errorf("failed to create spot order for %s", ticker)
	}

	return &models.ExchangeOrderDTO{
		ExchangeOrderID: order.OrderID,
		ClientOrderID:   req.NewClientOrderID,
		Amount:          amount,
	}, nil
}

func roundToSizeStep(amount decimal.Decimal, sizeStep string, assetPrecision int) decimal.Decimal {
	step, err := decimal.NewFromString(sizeStep)
	if err != nil || !step.IsPositive() {
		return amount.RoundDown(int32(assetPrecision)) //nolint:gosec
	}
	return amount.Div(step).Floor().Mul(step)
}

func isSideAllowed(symbol responses.SymbolInfo, side string) bool {
	switch symbol.TradeSideType {
	case mexc.TradeSideTypeAll:
		return true
	case mexc.TradeSideTypeBuyOnly:
		return side == mexc.OrderSideBuy
	case mexc.TradeSideTypeSellOnly:
		return side == mexc.OrderSideSell
	default:
		return false
	}
}

func (o *Service) GetOrderRule(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	info, err := o.exClient.Market().GetExchangeInfo(ctx, ticker)
	if err != nil {
		return nil, fmt.Errorf("get order rule for %s: %w", ticker, err)
	}
	if len(info.Symbols) == 0 {
		return nil, fmt.Errorf("symbol %s not found", ticker)
	}

	symbol := info.Symbols[0]
	if !isTradable(symbol) {
		return nil, exchangeclient.ErrSymbolTradingHalted
	}

	return &models.OrderRulesDTO{
		Symbol:                 symbol.Symbol,
		BaseCurrency:           symbol.BaseAsset,
		QuoteCurrency:          symbol.QuoteAsset,
		MinOrderAmount:         symbol.BaseSizePrecision,
		MinOrderValue:          cmp.Or(symbol.QuoteAmountPrecisionMarket, symbol.QuoteAmountPrecision),
		BuyMarketMaxOrderValue: cmp.Or(symbol.MaxQuoteAmountMarket, symbol.MaxQuoteAmount),
		AmountPrecision:        symbol.BaseAssetPrecision,
		PricePrecision:         symbol.QuotePrecision,
		ValuePrecision:         symbol.QuoteAssetPrecision,
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

func (o *Service) GetOrderDetails(ctx context.Context, args *models.GetOrderByIDParams) (*models.OrderDetailsDTO, error) {
	order := &models.OrderDetailsDTO{
		State:     models.ExchangeOrderStatusFailed,
		Amount:    decimal.Zero,
		AmountUSD: decimal.Zero,
	}

	if args.InstrumentID == nil || *args.InstrumentID == "" {
		return order, fmt.Errorf("instrument id is required")
	}

	req := &mexcrequests.QueryOrderRequest{Symbol: *args.InstrumentID}
	switch {
	case args.ExternalOrderID != nil && *args.ExternalOrderID != "":
		req.OrderID = *args.ExternalOrderID
	case args.ClientOrderID != nil && *args.ClientOrderID != "":
		req.OrigClientOrderID = *args.ClientOrderID
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
	case responses.OrderStatusNew, responses.OrderStatusPartiallyFilled:
		order.State = models.ExchangeOrderStatusInProgress
	case responses.OrderStatusCanceled, responses.OrderStatusPartiallyCanceled:
		order.State = models.ExchangeOrderStatusFailed
	default:
		order.State = models.ExchangeOrderStatusInProgress
	}

	order.Amount = res.ExecutedQty

	rule, err := o.GetOrderRule(ctx, res.Symbol)
	if err != nil {
		return nil, fmt.Errorf("get order rule for %s: %w", res.Symbol, err)
	}

	if rule.QuoteCurrency == models.CurrencyCodeUSDT || rule.QuoteCurrency == models.CurrencyCodeUSDC {
		order.AmountUSD = res.CummulativeQuoteQty
		return order, nil
	}

	amountUSD, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
		Source: models.ExchangeSlugMexc.String(),
		From:   rule.QuoteCurrency,
		To:     models.CurrencyCodeUSDT,
		Amount: res.CummulativeQuoteQty.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("convert %s to usdt: %w", rule.QuoteCurrency, err)
	}
	order.AmountUSD = amountUSD.Round(4)

	return order, nil
}

func (o *Service) GetWithdrawalRules(ctx context.Context, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	currEnabled, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugMexc)
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
				return item.Ticker == coin.Coin && item.Chain == network.NetWork
			}) {
				continue
			}

			minDepositAmount, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
				Source:     models.ExchangeSlugMexc.String(),
				From:       models.CurrencyCodeUSDT,
				To:         coin.Coin,
				Amount:     "1",
				StableCoin: false,
			})
			if err != nil {
				return nil, fmt.Errorf("convert minimum deposit amount for %s: %w", coin.Coin, err)
			}

			rules = append(rules, &models.WithdrawalRulesDTO{
				Currency:           coin.Coin,
				Chain:              network.NetWork,
				MinDepositAmount:   minDepositAmount.String(),
				MinWithdrawAmount:  network.WithdrawMin.String(),
				MaxWithdrawAmount:  network.WithdrawMax.String(),
				NumOfConfirmations: strconv.FormatInt(network.MinConfirm, 10),
				WithdrawPrecision:  network.WithdrawIntegerMultiple.String(),
				WithdrawFeeType:    models.WithdrawalFeeTypeFixed,
				Fee:                network.WithdrawFee.String(),
			})
		}
	}

	return rules, nil
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

	history, err := o.exClient.Wallet().GetWithdrawHistory(ctx, &mexcrequests.GetWithdrawHistoryRequest{})
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
