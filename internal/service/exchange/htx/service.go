package htx

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_chains"
	"github.com/dv-net/dv-merchant/internal/tools/hash"
	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/htx"
	htxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/models"
	htxrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/requests"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/ulule/limiter/v3"
)

var (
	ErrInsufficientBalance           = errors.New("insufficient balance")
	ErrUnprocessableCurrencyState    = errors.New("unprocessable currency state")
	ErrTradingDisabled               = errors.New("trading is disabled")
	ErrMaxOrderValueReached          = errors.New("max order value reached")
	ErrGetAccountBalanceZeroAccounts = errors.New("zero accounts returned")
)

const (
	WithdrawalStep = 10
)

type Service struct {
	exClient *htx.BaseClient
	storage  storage.IStorage
	convSvc  currconv.ICurrencyConvertor
	l        logger.Logger
	connHash string
}

func (o *Service) GetOrderDetails(ctx context.Context, args *models.GetOrderByIDParams) (*models.OrderDetailsDTO, error) {
	order := &models.OrderDetailsDTO{
		State:     models.ExchangeOrderStatusFailed,
		Amount:    decimal.Zero,
		AmountUSD: decimal.Zero,
	}
	if args.ExternalOrderID != nil {
		return o.getOrderDetailsExternal(ctx, order, *args.ExternalOrderID)
	}
	if args.ClientOrderID != nil {
		return o.getOrderDetailsInternal(ctx, order, *args.ClientOrderID)
	}
	return order, nil
}

func (o *Service) getOrderDetailsExternal(ctx context.Context, order *models.OrderDetailsDTO, externalOrderID string) (*models.OrderDetailsDTO, error) {
	o.l.Info("retrieving external order details", "exchange_slug", models.ExchangeSlugHtx, "external_order_id", externalOrderID, "connection_hash", o.connHash)

	orderID, err := strconv.ParseInt(externalOrderID, 10, 64)
	if err != nil {
		o.l.Error("failed to parse external order id", err, "exchange_slug", models.ExchangeSlugHtx, "external_order_id", externalOrderID, "connection_hash", o.connHash)
		return nil, fmt.Errorf("order id cant be casted to int64 %w", err)
	}
	res, err := o.exClient.Order().GetOrderDetails(ctx, orderID)
	if err != nil && errors.Is(err, htxmodels.ErrHtxBaseRecordInvalid) {
		o.l.Info("order not found on exchange", "exchange_slug", models.ExchangeSlugHtx, "external_order_id", externalOrderID, "connection_hash", o.connHash)
		return order, nil
	}
	if err != nil {
		o.l.Error("failed to get order details from exchange", err, "exchange_slug", models.ExchangeSlugHtx, "external_order_id", externalOrderID, "connection_hash", o.connHash)
		return nil, err
	}
	if res.Order != nil {
		amt, err := decimal.NewFromString(res.Order.Amount)
		if err != nil {
			o.l.Error("failed to parse order amount", err, "exchange_slug", models.ExchangeSlugHtx, "external_order_id", externalOrderID, "amount", res.Order.Amount, "connection_hash", o.connHash)
			return nil, err
		}
		order.Amount = amt

		amtUSD, err := o.getNotionalUSD(ctx, res.Order.Symbol, res.Order.Type, amt)
		if err != nil {
			o.l.Error("failed to calculate usd amount", err, "exchange_slug", models.ExchangeSlugHtx, "external_order_id", externalOrderID, "symbol", res.Order.Symbol, "connection_hash", o.connHash)
			return nil, err
		}
		order.AmountUSD = amtUSD

		switch res.Order.State {
		case htxmodels.OrderStateFilled:
			order.State = models.ExchangeOrderStatusCompleted
		default:
			order.State = models.ExchangeOrderStatusInProgress
		}

		o.l.Info("order details retrieved successfully", "exchange_slug", models.ExchangeSlugHtx, "external_order_id", externalOrderID, "state", res.Order.State.String(), "amount", amt.String(), "amount_usd", amtUSD.String(), "connection_hash", o.connHash)
	} else {
		o.l.Info("order response contains no order data", "exchange_slug", models.ExchangeSlugHtx, "external_order_id", externalOrderID, "connection_hash", o.connHash)
	}
	return order, nil
}

func (o *Service) getNotionalUSD(ctx context.Context, instID string, side htxmodels.OrderType, amount decimal.Decimal) (decimal.Decimal, error) {
	res, err := o.exClient.Common().GetAllSupportedSymbols(ctx)
	if err != nil {
		return decimal.Zero, err
	}
	symbol, exists := lo.Find(res.Symbols, func(s *htxmodels.Symbol) bool {
		return s.OutsideSymbol == instID
	})
	if !exists {
		return decimal.Zero, fmt.Errorf("symbol %s not found", instID)
	}
	currency := symbol.BaseCurrency
	if side == htxmodels.OrderTypeBuyMarket {
		currency = symbol.QuoteCurrency
	}
	if currency == "usdt" {
		return amount, nil
	}
	tickers, err := o.exClient.Market().GetMarketTickers(ctx)
	if err != nil {
		return decimal.Zero, err
	}
	ticker, exists := lo.Find(tickers.Tickers, func(t *htxmodels.MarketTicker) bool {
		return t.Symbol == currency+"usdt"
	})
	if !exists {
		return decimal.Zero, fmt.Errorf("ticker %s not found", currency+"usdt")
	}
	amt := decimal.NewFromFloat(ticker.Ask)
	return amount.Mul(amt), nil
}

func (o *Service) getOrderDetailsInternal(ctx context.Context, order *models.OrderDetailsDTO, clientOrderID string) (*models.OrderDetailsDTO, error) {
	o.l.Info("retrieving internal order details", "exchange_slug", models.ExchangeSlugHtx, "client_order_id", clientOrderID, "connection_hash", o.connHash)

	res, err := o.exClient.Order().GetOrderDetailsByClientID(ctx, &htxrequests.GetOrderByClientIDRequest{
		ClientOrderID: clientOrderID,
	})
	if err != nil && errors.Is(err, htxmodels.ErrHtxBaseRecordInvalid) {
		o.l.Info("order not found on exchange", "exchange_slug", models.ExchangeSlugHtx, "client_order_id", clientOrderID, "connection_hash", o.connHash)
		return order, nil
	}
	if err != nil {
		o.l.Error("failed to get order details by client id", err, "exchange_slug", models.ExchangeSlugHtx, "client_order_id", clientOrderID, "connection_hash", o.connHash)
		return nil, err
	}
	if res.Order != nil {
		amt, err := decimal.NewFromString(res.Order.Amount)
		if err != nil {
			o.l.Error("failed to parse order amount", err, "exchange_slug", models.ExchangeSlugHtx, "client_order_id", clientOrderID, "amount", res.Order.Amount, "connection_hash", o.connHash)
			return nil, err
		}
		order.Amount = amt
		switch res.Order.State {
		case htxmodels.OrderStateFilled:
			order.State = models.ExchangeOrderStatusCompleted
		default:
			order.State = models.ExchangeOrderStatusInProgress
		}

		o.l.Info("internal order details retrieved successfully", "exchange_slug", models.ExchangeSlugHtx, "client_order_id", clientOrderID, "state", res.Order.State.String(), "amount", amt.String(), "connection_hash", o.connHash)
	} else {
		o.l.Info("order response contains no order data", "exchange_slug", models.ExchangeSlugHtx, "client_order_id", clientOrderID, "connection_hash", o.connHash)
	}
	return order, nil
}

func (o *Service) GetOrderRule(ctx context.Context, symbol string) (*models.OrderRulesDTO, error) {
	res, err := o.exClient.Common().GetAllMarketSymbols(ctx, &htxrequests.GetMarketSymbolsRequest{Symbols: symbol})
	if err != nil {
		return nil, err
	}
	if len(res.MarketSymbols) == 0 { // redundant check
		return nil, fmt.Errorf("symbol not found")
	}
	s := res.MarketSymbols[0]

	return &models.OrderRulesDTO{
		Symbol:                   s.Symbol,
		State:                    s.State.String(),
		BaseCurrency:             s.BaseCurrency,
		QuoteCurrency:            s.QuoteCurrency,
		PricePrecision:           s.PricePrecision,
		AmountPrecision:          s.AmountPrecision,
		ValuePrecision:           s.ValuePrecision,
		MinOrderAmount:           decimal.NewFromFloat(s.MinOrderAmount).String(),
		MaxOrderAmount:           decimal.NewFromFloat(s.MaxOrderAmount).String(),
		MinOrderValue:            decimal.NewFromFloat(s.MinOrderValue).String(),
		SellMarketMinOrderAmount: decimal.NewFromFloat(s.SellMarketMinOrderAmount).String(),
		SellMarketMaxOrderAmount: decimal.NewFromFloat(s.SellMarketMaxOrderAmount).String(),
		BuyMarketMaxOrderValue:   decimal.NewFromFloat(s.BuyMarketMaxOrderValue).String(),
	}, nil
}

func (o *Service) GetCurrencyBalance(ctx context.Context, currency string) (*decimal.Decimal, error) {
	account, err := o.getSpotAccount(ctx)
	if err != nil {
		return nil, err
	}

	res, err := o.exClient.Account().GetAccountBalance(ctx, account)
	if err != nil {
		return nil, err
	}

	for _, b := range res.Balances.List {
		if b.Type == htxmodels.BalanceTypeTrade && b.Balance != "0" && strings.EqualFold(b.Currency, currency) {
			amount, err := decimal.NewFromString(b.Balance)
			if err != nil {
				return nil, err
			}
			return &amount, nil
		}
	}

	return &decimal.Zero, nil
}

func (o *Service) GetWithdrawalRules(ctx context.Context, currencyIDs ...string) ([]*models.WithdrawalRulesDTO, error) {
	enabledCurrencies, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugHtx)
	if err != nil {
		return nil, err
	}

	// TODO: might not work
	enabledCurrencies = lo.Filter(enabledCurrencies, func(i *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) bool {
		return lo.Contains(currencyIDs, i.ID.String)
	})

	if len(enabledCurrencies) == 0 {
		return nil, fmt.Errorf("currencies %s are not supported", strings.Join(currencyIDs, ","))
	}

	currReferences := make([]*htxmodels.CurrencyReference, 0, len(enabledCurrencies))
	for _, curr := range enabledCurrencies {
		ref, err := o.exClient.Common().GetCurrencyReference(ctx, &htxrequests.GetCurrencyReferenceRequest{
			Currency: curr.Ticker,
		})
		if err != nil {
			return nil, err
		}
		currReferences = append(currReferences, ref.CurrencyReference...)
	}

	exchangeRules := lo.FilterMap(currReferences, func(item *htxmodels.CurrencyReference, _ int) (*models.WithdrawalRulesDTO, bool) {
		validChain := lo.FilterMap(item.Chains, func(chain htxmodels.CurrencyReferenceChain, _ int) (htxmodels.CurrencyReferenceChain, bool) {
			if lo.ContainsBy(enabledCurrencies, func(c *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
				return c.Chain == chain.Chain
			}) {
				return chain, true
			}
			return chain, false
		})
		if len(validChain) == 0 {
			return nil, false
		}
		return &models.WithdrawalRulesDTO{
			Currency:            item.Currency,
			Chain:               validChain[0].Chain,
			MinDepositAmount:    validChain[0].MinDepositAmt,
			MinWithdrawAmount:   validChain[0].MinWithdrawAmt,
			MaxWithdrawAmount:   validChain[0].MaxWithdrawAmt,
			NumOfConfirmations:  strconv.Itoa(validChain[0].NumOfConfirmations),
			WithdrawFeeType:     models.WithdrawalFeeType(validChain[0].WithdrawFeeType),
			WithdrawPrecision:   strconv.Itoa(validChain[0].WithdrawPrecision),
			WithdrawQuotaPerDay: validChain[0].WithdrawQuotaPerDay,
			Fee:                 validChain[0].TransactFeeWithdraw,
		}, true
	})

	return exchangeRules, nil
}

func (o *Service) CreateSpotOrder(ctx context.Context, _ string, _ string, side string, ticker string, amount *decimal.Decimal, _ *models.OrderRulesDTO) (*models.ExchangeOrderDTO, error) { //nolint:staticcheck
	o.l.Info("starting spot order creation", "exchange_slug", models.ExchangeSlugHtx, "side", side, "ticker", ticker, "connection_hash", o.connHash)

	spotAccount, err := o.getSpotAccount(ctx)
	if err != nil {
		o.l.Error("failed to get spot account", err, "exchange_slug", models.ExchangeSlugHtx, "side", side, "ticker", ticker, "connection_hash", o.connHash)
		return nil, err
	}

	o.l.Info("retrieved spot account", "exchange_slug", models.ExchangeSlugHtx, "account_id", spotAccount.ID, "account_type", spotAccount.Type.String(), "side", side, "ticker", ticker, "connection_hash", o.connHash)

	balance, err := o.exClient.Account().GetAccountBalance(ctx, spotAccount)
	if err != nil {
		o.l.Error("failed to get account balance", err, "exchange_slug", models.ExchangeSlugHtx, "account_id", spotAccount.ID, "side", side, "ticker", ticker, "connection_hash", o.connHash)
		return nil, err
	}

	balances := lo.FilterMap(balance.Balances.List, func(item htxmodels.ListItem, _ int) (htxmodels.ListItem, bool) {
		if item.Balance != "0" && item.Type == htxmodels.BalanceTypeTrade {
			return item, true
		}
		return item, false
	})

	o.l.Info("filtered available balances", "exchange_slug", models.ExchangeSlugHtx, "balance_count", len(balances), "side", side, "ticker", ticker, "connection_hash", o.connHash)

	orderID, err := uuid.NewUUID()
	if err != nil {
		o.l.Error("failed to generate order id", err, "exchange_slug", models.ExchangeSlugHtx, "side", side, "ticker", ticker, "connection_hash", o.connHash)
		return nil, err
	}

	spotOrderRequest := &htxrequests.PlaceOrderRequest{
		AccountID:     strconv.FormatInt(spotAccount.ID, 10),
		Type:          side + "-" + "market",
		ClientOrderID: orderID.String(),
		// SelfMatchPrevent: 1, // prevent buying/selling to/from self
		Source: "spot-api",
	}

	res, err := o.exClient.Common().GetAllMarketSymbols(ctx, &htxrequests.GetMarketSymbolsRequest{
		Symbols: ticker,
	})
	if err != nil {
		return nil, err
	}

	if len(res.MarketSymbols) == 0 {
		return nil, fmt.Errorf("symbol %s not found", ticker)
	}
	symbol := res.MarketSymbols[0]

	if symbol.State != htxmodels.SymbolStatusOnline {
		return nil, fmt.Errorf("%w: symbol is not online", ErrUnprocessableCurrencyState)
	}
	if symbol.APITrading != "enabled" {
		return nil, fmt.Errorf("%w: symbol is not enabled for trading", ErrTradingDisabled)
	}

	spotOrderRequest.Symbol = symbol.Symbol

	// fixme: MinOrderAmt is minimum order amount in base currency of pair (e.g. BTC for BTC/USDT)
	// fixme: MinOrderValue is minimum order amount in quote currency of pair (e.g. USDT for BTC/USDT)
	orderMinimumBase := decimal.NewFromFloat(symbol.MinOrderAmount)
	orderMinimumQuote := decimal.NewFromFloat(symbol.MinOrderValue)
	sellMarketMinOrderAmt := decimal.NewFromFloat(symbol.SellMarketMinOrderAmount)
	buyMarketMaxOrderValue := decimal.NewFromFloat(symbol.BuyMarketMaxOrderValue)

	maxAmount := decimal.Zero
	switch spotOrderRequest.Type {
	case htxmodels.OrderTypeSellMarket.String():
		maxAmount = o.getMaxAmount(symbol.BaseCurrency, balances)
		if maxAmount.LessThan(orderMinimumBase) && maxAmount.LessThan(sellMarketMinOrderAmt) {
			return nil, ErrInsufficientBalance
		}
	case htxmodels.OrderTypeBuyMarket.String():
		maxAmount = o.getMaxAmount(symbol.QuoteCurrency, balances)
		if maxAmount.LessThan(orderMinimumQuote) {
			return nil, ErrInsufficientBalance
		}
		if maxAmount.GreaterThan(buyMarketMaxOrderValue) {
			return nil, ErrMaxOrderValueReached
		}
	default:
		return nil, fmt.Errorf("unsupported order type %s", spotOrderRequest.Type)
	}

	amount = &maxAmount //nolint:staticcheck

	switch spotOrderRequest.Type {
	case htxmodels.OrderTypeSellMarket.String():
		spotOrderRequest.Amount = maxAmount.RoundDown(int32(symbol.AmountPrecision)).String() //nolint:gosec
	case htxmodels.OrderTypeBuyMarket.String():
		spotOrderRequest.Amount = maxAmount.RoundDown(int32(symbol.ValuePrecision)).String() //nolint:gosec
	}

	placeOrderResponse, err := o.exClient.Order().PlaceOrder(ctx, spotOrderRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	order := &models.ExchangeOrderDTO{
		ExchangeOrderID: placeOrderResponse.OrderID,
		ClientOrderID:   orderID.String(),
		Amount:          *amount,
	}

	return order, nil
}

func (o *Service) getMaxAmount(baseSymbol string, accountBalance []htxmodels.ListItem) decimal.Decimal {
	for _, b := range accountBalance {
		if strings.EqualFold(b.Currency, baseSymbol) {
			amount, err := decimal.NewFromString(b.Available)
			if err != nil {
				return decimal.Zero
			}
			return amount
		}
	}
	return decimal.Zero
}

func (o *Service) checkSufficientBalance(balance []*models.AccountBalanceDTO, baseSymbol string, amount decimal.Decimal, minOrderAmtUSD decimal.Decimal) error { //nolint:unused
	for _, b := range balance {
		if strings.EqualFold(b.Currency, baseSymbol) {
			if b.Amount.LessThan(amount) {
				return ErrInsufficientBalance
			}
			if b.AmountUSD.LessThan(minOrderAmtUSD) { // fixme: not true for all currencies
				return ErrInsufficientBalance
			}
			return nil
		}
	}
	return ErrInsufficientBalance
}

func (o *Service) GetDepositAddresses(ctx context.Context, currency, _ string) ([]*models.DepositAddressDTO, error) {
	res, err := o.exClient.Wallet().GetDepositAddress(ctx, &htxrequests.DepositAddressRequest{
		Currency: currency,
	})
	if err != nil || len(res.Addresses) == 0 {
		return nil, err
	}

	addresses := make([]*models.DepositAddressDTO, 0, len(res.Addresses))
	for _, address := range res.Addresses {
		currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByParams(ctx, repo_exchange_chains.GetCurrencyIDByParamsParams{
			Ticker: currency,
			Chain:  address.Chain,
			Slug:   models.ExchangeSlugHtx,
		})
		if err != nil && errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get internal currency id %s: %w", address.Chain, err)
		}
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, &models.DepositAddressDTO{
			Address:          address.Address,
			Currency:         currencyID,
			Chain:            address.Chain,
			AddressType:      models.DepositAddress,
			InternalCurrency: address.Currency,
		})
	}
	return addresses, nil
}

func (o *Service) CreateWithdrawalOrder(ctx context.Context, args *models.CreateWithdrawalOrderParams) (*models.ExchangeWithdrawalDTO, error) {
	args.NativeAmount = args.NativeAmount.RoundDown(int32(args.WithdrawalPrecision)) //nolint:gosec

	internalCurrency, err := o.storage.ExchangeChains().GetTickerByCurrencyID(ctx, repo_exchange_chains.GetTickerByCurrencyIDParams{
		CurrencyID: args.Currency,
		Slug:       models.ExchangeSlugHtx,
	})
	if err != nil {
		return nil, err
	}

	req := &htxrequests.WithdrawVirtualCurrencyRequest{
		Address:  args.Address,
		Currency: internalCurrency,
		Chain:    args.Chain,
		Fee:      args.Fee.String(),
	}

	clientOrderID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	req.ClientOrderID = strings.ReplaceAll(clientOrderID.String(), "-", "")

	internalCurrencyID, err := o.storage.ExchangeChains().GetTickerByCurrencyID(ctx, repo_exchange_chains.GetTickerByCurrencyIDParams{
		CurrencyID: args.Currency,
		Slug:       models.ExchangeSlugHtx,
	})
	if err != nil {
		return nil, err
	}

	amount := args.NativeAmount.Sub(args.Fee)
	minWithdrawal := args.MinWithdrawal

	dto := &models.ExchangeWithdrawalDTO{}

	for {
		if amount.LessThan(minWithdrawal) {
			o.l.Info("withdrawal amount below minimum",
				"exchange", models.ExchangeSlugHtx.String(),
				"recordID", args.RecordID.String(),
				"current_amount", amount.String(),
				"min_withdrawal", minWithdrawal.String(),
			)
			return nil, exchangeclient.ErrMinWithdrawalBalance
		}

		withdrawalStep, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
			Source:     models.ExchangeSlugHtx.String(),
			From:       "USDT",
			To:         internalCurrencyID,
			Amount:     decimal.NewFromInt(WithdrawalStep).String(),
			StableCoin: false,
		})
		if err != nil {
			return nil, err
		}

		req.Amount = amount.String()

		res, err := o.exClient.Wallet().WithdrawVirtualCurrency(ctx, req)
		if err != nil {
			if errors.Is(err, exchangeclient.ErrWithdrawalBalanceLocked) {
				o.l.Error("insufficient funds, retrying with reduced amount",
					exchangeclient.ErrWithdrawalBalanceLocked,
					"exchange", models.ExchangeSlugHtx.String(),
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
		dto.InternalOrderID = req.ClientOrderID
		dto.ExternalOrderID = strconv.FormatInt(res.WithdrawalTransferID, 10)
		return dto, nil
	}
}

func NewService(l logger.Logger, accessKey, secretKey string, baseURL *url.URL, storage storage.IStorage, store limiter.Store, convSvc currconv.ICurrencyConvertor) (*Service, error) {
	exClient, err := htx.NewBaseClient(&htx.ClientOptions{
		AccessKey: accessKey,
		SecretKey: secretKey,
		BaseURL:   baseURL,
	}, store, htx.WithLogger(l))
	if err != nil {
		return nil, err
	}
	connHash, err := hash.SHA256ConnectionHash(models.ExchangeSlugHtx.String(), accessKey, secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection hash: %w", err)
	}

	return &Service{
		exClient: exClient,
		storage:  storage,
		convSvc:  convSvc,
		l:        l,
		connHash: connHash,
	}, nil
}

func (o *Service) GetAccountBalance(ctx context.Context) ([]*models.AccountBalanceDTO, error) {
	enabledCurrencies, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugHtx)
	if err != nil {
		return nil, err
	}

	account, err := o.getSpotAccount(ctx)
	if err != nil {
		return nil, err
	}

	res, err := o.exClient.Account().GetAccountBalance(ctx, account)
	if err != nil {
		return nil, err
	}

	var dto []*models.AccountBalanceDTO
	for _, b := range res.Balances.List {
		if slices.IndexFunc(enabledCurrencies, func(c *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
			return c.Ticker == b.Currency
		}) == -1 {
			continue
		}
		if b.Type == htxmodels.BalanceTypeTrade && b.Balance != "0" {
			amount, err := decimal.NewFromString(b.Balance)
			if err != nil {
				return nil, err
			}
			amountUSD, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
				Source:     models.ExchangeSlugHtx.String(),
				From:       b.Currency,
				To:         "USDT",
				Amount:     amount.String(),
				StableCoin: false,
			})
			if err != nil {
				return nil, err
			}
			currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByTicker(ctx, b.Currency)
			if err != nil {
				continue
			}
			dto = append(dto, &models.AccountBalanceDTO{
				Currency:  currencyID,
				Type:      models.CurrencyTypeCrypto.String(),
				Amount:    amount,
				AmountUSD: amountUSD.Round(4),
			})
		}
	}

	return dto, nil
}

func (o *Service) GetWithdrawalByID(ctx context.Context, args *models.GetWithdrawalByIDParams) (*models.WithdrawalStatusDTO, error) {
	request := &htxrequests.WithdrawalByClientIDRequest{}
	if args.ClientOrderID == nil {
		return nil, fmt.Errorf("ClientOrderID is required")
	}
	if args.ClientOrderID != nil {
		request.ClientOrderID = *args.ClientOrderID
	}

	wdRecord, err := o.exClient.Wallet().GetWithdrawalByClientID(ctx, request)
	if err != nil {
		return nil, err
	}
	if wdRecord.WithdrawalTransferData == nil {
		return nil, fmt.Errorf("withdrawal not found")
	}
	res := &models.WithdrawalStatusDTO{
		ID:     strconv.Itoa(wdRecord.WithdrawalTransferData.ID),
		Status: wdRecord.WithdrawalTransferData.State.String(),
	}
	if wdRecord.WithdrawalTransferData.TxHash != "" {
		res.TxHash = wdRecord.WithdrawalTransferData.TxHash
	}
	if wdRecord.ErrMsg != "" {
		res.ErrorMessage = wdRecord.ErrMsg
	}
	if wdRecord.WithdrawalTransferData.Amount != 0.0 {
		res.NativeAmount = decimal.NewFromFloat(wdRecord.WithdrawalTransferData.Amount)
	}
	return res, nil
}

func (o *Service) TestConnection(ctx context.Context) error {
	res, err := o.exClient.User().GetUserUID(ctx)
	if err != nil {
		return err
	}
	if res.Data != 0 {
		// Retrieve API keys
		info, err := o.exClient.User().GetAPIKeyInformation(ctx, &htxrequests.GetAPIKeyInformationRequest{
			UID:       strconv.Itoa(int(res.Data)),
			AccessKey: o.exClient.AccessKey(),
		})
		if err != nil {
			return err
		}
		// Find specific key used in this client
		apiInfo, exists := lo.Find(info.Data, func(entry *htxmodels.APIKeyInformation) bool {
			return entry.AccessKey == o.exClient.AccessKey()
		})
		if exists {
			perms := strings.Split(apiInfo.Permission, ",")
			requiredPerms := []string{"readOnly", "trade", "withdraw"}
			for _, required := range requiredPerms {
				if !lo.Contains(perms, required) {
					return exchangeclient.ErrIncorrectAPIPermissions
				}
			}
		}
	}
	return nil
}

func (o *Service) GetExchangeSymbols(ctx context.Context) ([]*models.ExchangeSymbolDTO, error) {
	res, err := o.exClient.Common().GetAllMarketSymbols(ctx, &htxrequests.GetMarketSymbolsRequest{})
	if err != nil {
		return nil, err
	}

	dto := make([]*models.ExchangeSymbolDTO, 0, len(res.MarketSymbols)*2)
	for _, s := range res.MarketSymbols {
		if s.State != htxmodels.SymbolStatusOnline {
			continue
		}
		base, quote := strings.ToUpper(s.BaseCurrency), strings.ToUpper(s.QuoteCurrency)
		dto = append(dto, &models.ExchangeSymbolDTO{
			Symbol:      s.Symbol,
			DisplayName: base + "/" + quote,
			BaseSymbol:  s.BaseCurrency,
			QuoteSymbol: s.QuoteCurrency,
			Type:        "sell",
		}, &models.ExchangeSymbolDTO{
			Symbol:      s.Symbol,
			DisplayName: quote + "/" + base,
			BaseSymbol:  s.BaseCurrency,
			QuoteSymbol: s.QuoteCurrency,
			Type:        "buy",
		})
	}

	return dto, nil
}

func (o *Service) getSpotAccount(ctx context.Context) (*htxmodels.Account, error) {
	res, err := o.exClient.Account().GetAllAccounts(ctx)
	if err != nil {
		return nil, err
	}
	if len(res.Accounts) == 0 {
		return nil, ErrGetAccountBalanceZeroAccounts
	}
	firstSpotAcc := slices.IndexFunc(res.Accounts, func(acc *htxmodels.Account) bool {
		return acc.Type == htxmodels.AccountTypeSpot && acc.State == htxmodels.AccountStateWorking
	})
	if firstSpotAcc == -1 {
		return nil, fmt.Errorf("failed to find working spot account")
	}

	return res.Accounts[firstSpotAcc], nil
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
