package kucoin

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
	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
	"github.com/dv-net/dv-merchant/pkg/key_value"
	"github.com/dv-net/dv-merchant/pkg/logger"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	kucoinclients "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin"
	kucoinmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/models"
	kucoinrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/requests"
)

const (
	WithdrawalStep = 10
)

var (
	ErrInsufficientBalance         = errors.New("insufficient balance")
	ErrUnprocessableCurrencyStatus = errors.New("unprocessable currency status")
	ErrMaxOrderValueReached        = errors.New("max order value reached")
)

func NewService(l logger.Logger, accessKey, secretKey, passPhrase string, _ bool, baseURL *url.URL, storage storage.IStorage, store limiter.Store, convSvc currconv.ICurrencyConvertor) (*Service, error) {
	// Create cache for public endpoint caching
	cache := key_value.NewInMemory()

	exClient := kucoinclients.NewBaseClient(&kucoinclients.ClientOptions{
		KeyAPI:        accessKey,
		KeySecret:     secretKey,
		KeyPassphrase: passPhrase,
		BaseURL:       baseURL,
	}, store, cache, kucoinclients.WithLogger(l))

	connHash, err := hash.SHA256ConnectionHash(models.ExchangeSlugKucoin.String(), accessKey, secretKey, passPhrase)
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

type Service struct {
	exClient kucoinclients.IKucoinClient
	storage  storage.IStorage
	convSvc  currconv.ICurrencyConvertor
	l        logger.Logger
	connHash string
}

func (o *Service) TestConnection(ctx context.Context) error {
	res, err := o.exClient.Account().GetAPIKeyInfo(ctx, kucoinrequests.GetAPIKeyInfo{})
	if err != nil {
		return err
	}
	// Test if permission string contains "General", "Spot", "Transfer", "InnerTransfer"
	perms := strings.Split(res.Info.Permission, ",")
	requiredPerms := []string{"General", "Spot", "InnerTransfer", "Transfer"}
	for _, required := range requiredPerms {
		if !lo.Contains(perms, required) {
			return exchangeclient.ErrIncorrectAPIPermissions
		}
	}
	return nil
}

func (o *Service) GetAccountBalance(ctx context.Context) ([]*models.AccountBalanceDTO, error) {
	// Fetch all account's balances
	res, err := o.exClient.Account().GetAccountList(ctx, kucoinrequests.GetAccountList{})
	if err != nil {
		return nil, err
	}
	// Filter out the accounts that are not "main" or "trading"
	accounts := lo.Filter(res.Accounts, func(item *kucoinmodels.Account, _ int) bool {
		return item.Type == kucoinmodels.AccountTypeMain || item.Type == kucoinmodels.AccountTypeTrade
	})
	if len(accounts) == 0 { // no accounts found - being zero balance
		return []*models.AccountBalanceDTO{}, nil
	}

	// Iterate over the accounts and sum up the available balances
	balances := map[string]decimal.Decimal{}
	for _, account := range accounts {
		if !account.Available.IsZero() {
			balances[account.Currency] = balances[account.Currency].Add(account.Available)
		}
	}
	// Convert the balances to AccountBalanceDTO
	accountBalances := make([]*models.AccountBalanceDTO, 0, len(balances))
	for currency, balance := range balances {
		currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByTicker(ctx, currency)
		if err != nil {
			continue
		}
		amountUSD, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
			Source:     "kucoin",
			From:       currency,
			To:         "USDT",
			Amount:     balance.String(),
			StableCoin: false,
		})
		if err != nil {
			return nil, err
		}

		accountBalances = append(accountBalances, &models.AccountBalanceDTO{
			Currency:  currencyID,
			Amount:    balance,
			AmountUSD: amountUSD.Round(4),
			Type:      models.CurrencyTypeCrypto.String(),
		})
	}

	return accountBalances, nil
}

func (o *Service) GetCurrencyBalance(ctx context.Context, currency string) (*decimal.Decimal, error) {
	// Fetch all account's balances
	res, err := o.exClient.Account().GetAccountList(ctx, kucoinrequests.GetAccountList{Currency: currency})
	if err != nil {
		return nil, err
	}
	// Filter out the accounts that are not "main" or "trading"
	accounts := lo.Filter(res.Accounts, func(item *kucoinmodels.Account, _ int) bool {
		return item.Type == kucoinmodels.AccountTypeMain || item.Type == kucoinmodels.AccountTypeTrade
	})
	if len(accounts) == 0 { // no accounts found - being zero balance
		return &decimal.Decimal{}, nil
	}
	// Iterate over the accounts and sum up the available balances
	balance := decimal.Decimal{}
	for _, account := range accounts {
		if !account.Available.IsZero() {
			balance = balance.Add(account.Available)
		}
	}
	return &balance, nil
}

func (o *Service) GetExchangeSymbols(ctx context.Context) ([]*models.ExchangeSymbolDTO, error) {
	// Fetch all exchange symbols (ETFs, etc.)
	res, err := o.exClient.Public().GetAllSymbols(ctx, kucoinrequests.GetAllSymbols{})
	if err != nil {
		return nil, err
	}

	// Filter out the symbols that are paired with stable coins and BTC, and ALTs.
	// Exclude symbols featured for marginal only, or ETF trading
	// exchangeSymbols := make([]*models.ExchangeSymbolDTO, 0, len(res.Symbols)*2)
	res.Symbols = lo.Filter(res.Symbols, func(symbol *kucoinmodels.Symbol, _ int) bool {
		if symbol.EnableTrading &&
			(symbol.Market == kucoinmodels.MarketTypeALTS.String() ||
				symbol.Market == kucoinmodels.MarketTypeUSDS.String() ||
				symbol.Market == kucoinmodels.MarketTypeBTC.String()) &&
			!isLeveragedToken(symbol.BaseCurrency) &&
			!isLeveragedToken(symbol.QuoteCurrency) {
			return true
		}
		return false
	})

	exchangeSymbols := make([]*models.ExchangeSymbolDTO, 0, len(res.Symbols)*2)
	for _, symbol := range res.Symbols {
		baseSymbol, quoteSymbol := strings.ToUpper(symbol.BaseCurrency), strings.ToUpper(symbol.QuoteCurrency)
		exchangeSymbols = append(exchangeSymbols, &models.ExchangeSymbolDTO{
			Symbol:      symbol.Symbol,
			DisplayName: baseSymbol + "/" + quoteSymbol,
			BaseSymbol:  symbol.BaseCurrency,
			QuoteSymbol: symbol.QuoteCurrency,
			Type:        "sell",
		}, &models.ExchangeSymbolDTO{
			Symbol:      symbol.Symbol,
			DisplayName: quoteSymbol + "/" + baseSymbol,
			BaseSymbol:  symbol.BaseCurrency,
			QuoteSymbol: symbol.QuoteCurrency,
			Type:        "buy",
		})
	}
	return exchangeSymbols, nil
}

func isLeveragedToken(currency string) bool {
	return strings.HasSuffix(currency, "3L") ||
		strings.HasSuffix(currency, "2L") ||
		strings.HasSuffix(currency, "3S") ||
		strings.HasSuffix(currency, "2S")
}

func (o *Service) GetDepositAddresses(ctx context.Context, currency string, network string) ([]*models.DepositAddressDTO, error) {
	currencies, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugKucoin)
	if err != nil {
		return nil, err
	}
	currencies = lo.Filter(currencies, func(item *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) bool {
		return item.Ticker == currency && item.Chain == network
	})

	deposits := make([]*models.DepositAddressDTO, 0, len(currencies))
	for _, curr := range currencies {
		depositAddress := &models.DepositAddressDTO{}
		currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByParams(ctx, repo_exchange_chains.GetCurrencyIDByParamsParams{
			Ticker: currency,
			Chain:  curr.Chain,
			Slug:   models.ExchangeSlugKucoin,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get internal currency id %s: %w", network, err)
		}

		res, err := o.exClient.Account().GetDepositAddress(ctx, kucoinrequests.GetDepositAddress{
			Currency: currency,
			Chain:    network,
		})

		switch {
		case err != nil && (strings.Contains(err.Error(), "deposit.disabled") || strings.Contains(err.Error(), "111001")):
			// Create deposit address if it's disabled/not created
			createRes, createErr := o.exClient.Account().CreateDepositAddress(ctx, kucoinrequests.CreateDepositAddress{
				Currency: currency,
				Chain:    network,
			})
			if createErr != nil {
				return nil, createErr
			}
			depositAddress = &models.DepositAddressDTO{
				Address:          createRes.Address.Address,
				InternalCurrency: createRes.Address.Currency,
				Chain:            createRes.Address.ChainID,
				AddressType:      models.DepositAddress,
			}
		case err != nil:
			// If there's an error but it's not about deposit being disabled, return it
			return nil, err
		default:
			// No error and addresses exist, use existing address
			if len(res.Addresses) > 0 && res.Addresses[0] != nil {
				addressInfo := res.Addresses[0]
				depositAddress = &models.DepositAddressDTO{
					Address:          addressInfo.Address,
					InternalCurrency: addressInfo.Currency,
					Chain:            addressInfo.ChainID,
					AddressType:      models.DepositAddress,
				}
			}
		}

		depositAddress.Currency = currencyID
		deposits = append(deposits, depositAddress)
	}

	return deposits, nil
}

func (o *Service) CreateWithdrawalOrder(ctx context.Context, args *models.CreateWithdrawalOrderParams) (*models.ExchangeWithdrawalDTO, error) {
	precision := int32(args.WithdrawalPrecision)

	args.NativeAmount = args.NativeAmount.RoundDown(precision).Sub(args.NativeAmount.Div(decimal.NewFromInt(100)).Mul(decimal.NewFromInt(1))).RoundDown(precision)

	internalCurrencyID, err := o.storage.ExchangeChains().GetTickerByCurrencyID(ctx, repo_exchange_chains.GetTickerByCurrencyIDParams{
		CurrencyID: args.Currency,
		Slug:       models.ExchangeSlugKucoin,
	})
	if err != nil {
		return nil, err
	}

	res, err := o.exClient.Account().GetAccountList(ctx, kucoinrequests.GetAccountList{})
	if err != nil {
		return nil, err
	}

	fundingAccounts := lo.Filter(res.Accounts, func(item *kucoinmodels.Account, _ int) bool {
		return item.Type == kucoinmodels.AccountTypeMain
	})

	spotAccounts := lo.Filter(res.Accounts, func(item *kucoinmodels.Account, _ int) bool {
		return item.Type == kucoinmodels.AccountTypeTrade
	})

	fundingBalances := map[string]decimal.Decimal{}
	for _, account := range fundingAccounts {
		if !account.Available.IsZero() && account.Currency == internalCurrencyID {
			fundingBalances[account.Currency] = fundingBalances[account.Currency].Add(account.Available)
		}
	}

	spotBalances := map[string]decimal.Decimal{}
	for _, account := range spotAccounts {
		if !account.Available.IsZero() && account.Currency == internalCurrencyID {
			spotBalances[account.Currency] = spotBalances[account.Currency].Add(account.Available)
		}
	}

	spotBalance := o.getBalance(internalCurrencyID, spotBalances)
	fundingBalance := o.getBalance(internalCurrencyID, fundingBalances)

	totalBalance := spotBalance.Add(fundingBalance)

	o.l.Infow("balances",
		"exchange", models.ExchangeSlugKucoin.String(),
		"recordID", args.RecordID.String(),
		"withdrawalAmount", args.NativeAmount.String(),
		"withdrawalFee", args.Fee.String(),
		"totalBalance", totalBalance.String(),
		"spotBalance", spotBalance.String(),
		"fundingBalance", fundingBalance.String(),
	)

	if fundingBalance.LessThan(args.NativeAmount) {
		o.l.Infow("funding balance is less than withdrawal amount",
			"exchange", models.ExchangeSlugKucoin.String(),
			"recordID", args.RecordID.String(),
		)

		if totalBalance.LessThan(args.NativeAmount) {
			o.l.Infow("total balance is less than withdrawal amount",
				"exchange", models.ExchangeSlugKucoin.String(),
				"recordID", args.RecordID.String(),
			)
			return nil, exchangeclient.ErrMinWithdrawalBalance
		}

		transferOID, err := uuid.NewUUID()
		if err != nil {
			return nil, err
		}

		transferAmount := totalBalance.Sub(fundingBalance).RoundDown(precision)
		_, err = o.exClient.Account().CreateFlexTransfer(ctx, kucoinrequests.FlexTransfer{
			ClientOID:       transferOID.String(),
			Type:            kucoinmodels.TransferTypeInternal,
			Currency:        internalCurrencyID,
			Amount:          transferAmount.String(),
			FromAccountType: kucoinmodels.TransferAccountTypeTrade,
			ToAccountType:   kucoinmodels.TransferAccountTypeMain,
		})
		if err != nil {
			return nil, err
		}
		o.l.Infow("transfer funds from spot to funding",
			"exchange", models.ExchangeSlugKucoin.String(),
			"recordID", args.RecordID.String(),
			"transferOID", transferOID.String(),
			"transferAmount", transferAmount.String(),
			"fromAccountType", kucoinmodels.TransferAccountTypeTrade,
			"toAccountType", kucoinmodels.TransferAccountTypeMain,
		)
	}

	req := kucoinrequests.CreateWithdrawal{
		Currency:      internalCurrencyID,
		ToAddress:     args.Address,
		Chain:         args.Chain,
		WithdrawType:  kucoinmodels.WithdrawalTypeAddress,
		Amount:        args.NativeAmount.Sub(args.Fee).RoundDown(precision).String(),
		FeeDeductType: kucoinmodels.FeeDeductTypeExternal,
	}

	o.l.Infow("withdrawal request assembled",
		"exchange", models.ExchangeSlugKucoin.String(),
		"recordID", args.RecordID.String(),
		"withdrawalAmount", req.Amount,
		"withdrawalFee", args.Fee.String(),
		"withdrawalAddress", args.Address,
		"withdrawalChain", req.Chain,
		"withdrawalFeeType", req.FeeDeductType,
	)

	amount := args.NativeAmount.Sub(args.Fee).RoundDown(precision)
	minWithdrawals := args.MinWithdrawal

	dto := &models.ExchangeWithdrawalDTO{}
	for {
		if amount.LessThan(minWithdrawals) {
			o.l.Infow("withdrawal amount below minimum",
				"exchange", models.ExchangeSlugKucoin.String(),
				"recordID", args.RecordID.String(),
				"current_amount", amount.String(),
				"min_withdrawal", minWithdrawals.String(),
			)
			return nil, exchangeclient.ErrMinWithdrawalBalance
		}

		withdrawalStep, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
			Source:     models.ExchangeSlugKucoin.String(),
			From:       "USDT",
			To:         req.Currency,
			Amount:     decimal.NewFromInt(WithdrawalStep).String(),
			StableCoin: false,
		})
		if err != nil {
			return nil, err
		}

		req.Amount = amount.String()
		res, err := o.exClient.Account().CreateWithdrawal(ctx, req)
		if err == nil {
			dto.InternalOrderID = res.Data.WithdrawalID
			return dto, nil
		}

		if errors.Is(err, exchangeclient.ErrWithdrawalBalanceLocked) {
			o.l.Infow("withdrawal balance locked",
				"exchange", models.ExchangeSlugKucoin.String(),
				"recordID", args.RecordID.String(),
				"current_amount", amount.String(),
			)

			amount = amount.Sub(withdrawalStep).RoundDown(precision)
			if amount.LessThan(minWithdrawals) {
				return nil, exchangeclient.ErrMinWithdrawalBalance
			}
			dto.RetryReason = exchangeclient.ErrWithdrawalBalanceLocked.Error()
			continue
		}

		return nil, err
	}
}

func (o *Service) getBalance(symbol string, balances map[string]decimal.Decimal) decimal.Decimal {
	spotAmount := decimal.Zero
	if balance, exists := balances[symbol]; exists {
		spotAmount = spotAmount.Add(balance)
	}
	return spotAmount
}

func (o *Service) CreateSpotOrder(ctx context.Context, from string, to string, side string, ticker string, amount *decimal.Decimal, rule *models.OrderRulesDTO) (*models.ExchangeOrderDTO, error) { //nolint:all
	orderID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	tradingSymbol, err := o.exClient.Public().GetSymbol(ctx, kucoinrequests.GetSymbol{
		Symbol: ticker,
	})
	if err != nil {
		if errors.Is(err, exchangeclient.ErrRateLimited) {
			return nil, exchangeclient.ErrSkipOrder
		}
		return nil, err
	}

	if !tradingSymbol.Symbol.EnableTrading {
		return nil, fmt.Errorf("trading is disabled for symbol %s", ticker)
	}

	spotOrderRequest := kucoinrequests.CreateOrder{
		ClientOID: orderID.String(),
		Symbol:    ticker,
		Side:      kucoinmodels.OrderSide(side),
		Type:      kucoinmodels.OrderTypeMarket,
	}

	accountList, err := o.exClient.Account().GetAccountList(ctx, kucoinrequests.GetAccountList{})
	if err != nil {
		if errors.Is(err, exchangeclient.ErrRateLimited) {
			return nil, exchangeclient.ErrSkipOrder
		}
		return nil, err
	}
	spotAccounts := lo.Filter(accountList.Accounts, func(item *kucoinmodels.Account, _ int) bool {
		return item.Type == kucoinmodels.AccountTypeTrade
	})

	fundingAccounts := lo.Filter(accountList.Accounts, func(item *kucoinmodels.Account, _ int) bool {
		return item.Type == kucoinmodels.AccountTypeMain
	})

	spotBalances := map[string]decimal.Decimal{}
	for _, account := range spotAccounts {
		if !account.Available.IsZero() && account.Currency == tradingSymbol.Symbol.BaseCurrency || account.Currency == tradingSymbol.Symbol.QuoteCurrency {
			spotBalances[account.Currency] = spotBalances[account.Currency].Add(account.Available)
		}
	}

	fundingBalances := map[string]decimal.Decimal{}
	for _, account := range fundingAccounts {
		if !account.Available.IsZero() && account.Currency == tradingSymbol.Symbol.BaseCurrency || account.Currency == tradingSymbol.Symbol.QuoteCurrency {
			fundingBalances[account.Currency] = fundingBalances[account.Currency].Add(account.Available)
		}
	}

	maxAmount, spotAmount, fundingAmount := decimal.Zero, decimal.Zero, decimal.Zero //nolint:all

	orderMinimumBase, err := decimal.NewFromString(rule.MinOrderAmount)
	if err != nil {
		return nil, err
	}

	orderMinimumQuote, err := decimal.NewFromString(rule.MinOrderValue)
	if err != nil {
		return nil, err
	}

	fundsTransferRequest := kucoinrequests.FlexTransfer{
		ClientOID:       orderID.String(),
		Type:            kucoinmodels.TransferTypeInternal,
		FromAccountType: kucoinmodels.TransferAccountTypeMain,
		ToAccountType:   kucoinmodels.TransferAccountTypeTrade,
	}

	switch spotOrderRequest.Side {
	case kucoinmodels.OrderSideSell:
		fundsTransferRequest.Currency = tradingSymbol.Symbol.BaseCurrency
		spotAmount = o.getBalance(tradingSymbol.Symbol.BaseCurrency, spotBalances)
		fundingAmount = o.getBalance(tradingSymbol.Symbol.BaseCurrency, fundingBalances)
		maxAmount = spotAmount.Add(fundingAmount)
		if maxAmount.LessThan(orderMinimumBase) {
			return nil, exchangeclient.ErrInsufficientBalance
		}
	case kucoinmodels.OrderSideBuy:
		fundsTransferRequest.Currency = tradingSymbol.Symbol.QuoteCurrency
		spotAmount = o.getBalance(tradingSymbol.Symbol.QuoteCurrency, spotBalances)
		fundingAmount = o.getBalance(tradingSymbol.Symbol.QuoteCurrency, fundingBalances)
		maxAmount = spotAmount.Add(fundingAmount)
		if maxAmount.LessThan(orderMinimumQuote) {
			return nil, ErrInsufficientBalance
		}
	default:
		return nil, fmt.Errorf("unsupported order type %s", spotOrderRequest.Side)
	}

	remainingTopup := maxAmount.Sub(spotAmount)
	if remainingTopup.IsPositive() {
		remainder := remainingTopup.String()
		fundsTransferRequest.Amount = remainder

		_, err = o.exClient.Account().CreateFlexTransfer(ctx, fundsTransferRequest)
		if err != nil {
			if errors.Is(err, exchangeclient.ErrRateLimited) {
				return nil, exchangeclient.ErrSkipOrder
			}
			return nil, err
		}
	}

	amount = &maxAmount //nolint:staticcheck

	switch spotOrderRequest.Side {
	case kucoinmodels.OrderSideSell:
		maxAmount = maxAmount.Sub(maxAmount.Copy().Mul(decimal.NewFromFloat(0.001))).Div(tradingSymbol.Symbol.BaseIncrement).Floor().Mul(tradingSymbol.Symbol.BaseIncrement).RoundDown(int32(rule.AmountPrecision))
		spotOrderRequest.Size = maxAmount.String()
	case kucoinmodels.OrderSideBuy:
		maxAmount = maxAmount.Sub(maxAmount.Copy().Mul(decimal.NewFromFloat(0.001))).Div(tradingSymbol.Symbol.QuoteIncrement).Floor().Mul(tradingSymbol.Symbol.QuoteIncrement).RoundDown(int32(rule.ValuePrecision))
		spotOrderRequest.Funds = maxAmount.String()
	}

	// Validate order value before sending to exchange
	minOrderValue, err := decimal.NewFromString(rule.MinOrderValue)
	if err != nil {
		return nil, fmt.Errorf("invalid min order value: %w", err)
	}

	orderValueInQuote := maxAmount
	if spotOrderRequest.Side == kucoinmodels.OrderSideSell {
		tickerPrice, err := o.exClient.Public().GetTicker(ctx, kucoinrequests.GetTicker{
			Symbol: spotOrderRequest.Symbol,
		})
		if err != nil {
			if errors.Is(err, exchangeclient.ErrRateLimited) {
				return nil, exchangeclient.ErrSkipOrder
			}
			return nil, fmt.Errorf("failed to get ticker price: %w", err)
		}
		if tickerPrice.Ticker == nil || tickerPrice.Ticker.BestBid.IsZero() {
			return nil, fmt.Errorf("no valid ticker price available for %s: %w", spotOrderRequest.Symbol, exchangeclient.ErrSkipOrder)
		}
		orderValueInQuote = maxAmount.Mul(tickerPrice.Ticker.BestBid)
	}

	if orderValueInQuote.LessThan(minOrderValue) {
		o.l.Infow("order value below minimum after calculations, skipping",
			"exchange", models.ExchangeSlugKucoin.String(),
			"ticker", ticker,
			"side", side,
			"calculated_amount", maxAmount.String(),
			"calculated_value_quote", orderValueInQuote.String(),
			"min_required", minOrderValue.String(),
		)
		return nil, exchangeclient.ErrSkipOrder
	}

	placedOrder, err := o.exClient.Spot().CreateOrder(ctx, spotOrderRequest)
	if err != nil {
		if errors.Is(err, exchangeclient.ErrRateLimited) {
			return nil, exchangeclient.ErrSkipOrder
		}
		if errors.Is(err, exchangeclient.ErrMinOrderValue) {
			o.l.Infow("order value below minimum, skipping",
				"exchange", models.ExchangeSlugKucoin.String(),
				"ticker", ticker,
				"side", side,
				"amount", maxAmount.String(),
			)
			return nil, exchangeclient.ErrSkipOrder
		}
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	order := &models.ExchangeOrderDTO{
		ExchangeOrderID: placedOrder.Data.OrderID,
		ClientOrderID:   orderID.String(),
		Amount:          *amount,
	}

	return order, nil
}

func (o *Service) GetOrderRule(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	symbolData, err := o.exClient.Public().GetSymbol(ctx, kucoinrequests.GetSymbol{
		Symbol: ticker,
	})
	if err != nil {
		return nil, err
	}
	if symbolData.Symbol == nil {
		return nil, fmt.Errorf("symbol %s not found", ticker)
	}

	// Temporary fix for incorrect min size
	if symbolData.Symbol.BaseMinSize.LessThan(symbolData.Symbol.BaseIncrement) {
		symbolData.Symbol.BaseMinSize = symbolData.Symbol.BaseIncrement
	}
	if symbolData.Symbol.QuoteMinSize.LessThan(symbolData.Symbol.QuoteIncrement) {
		symbolData.Symbol.QuoteMinSize = symbolData.Symbol.QuoteIncrement
	}

	// baseIncrement represents order precision for base asset
	basePrecision := utils.ConvertPrecision(symbolData.Symbol.BaseIncrement.String())
	// quoteIncrement represents order precision for quote asset
	quotePrecision := utils.ConvertPrecision(symbolData.Symbol.QuoteIncrement.String())

	return &models.OrderRulesDTO{
		Symbol:          symbolData.Symbol.Symbol,
		BaseCurrency:    symbolData.Symbol.BaseCurrency,
		QuoteCurrency:   symbolData.Symbol.QuoteCurrency,
		MinOrderAmount:  symbolData.Symbol.BaseMinSize.String(),
		MinOrderValue:   symbolData.Symbol.MinFunds.Mul(decimal.NewFromInt(10)).String(), // KuCoin returns for ex. 0.1 USDT and we try 0.1161 LTC -> 0,105 USDT - fail. Temporary fix for this
		MaxOrderAmount:  symbolData.Symbol.BaseMaxSize.String(),
		AmountPrecision: int(basePrecision.IntPart()),
		ValuePrecision:  int(quotePrecision.IntPart()),
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
		State: models.ExchangeOrderStatusFailed,
	}
	// Since Kucoin does not return inactive orders older then 7 days
	if args.InternalOrder != nil {
		if args.InternalOrder.CreatedAt.Time.Before(time.Now().AddDate(0, 0, -7)) {
			o.l.Infow("order is older than 7 days, failing",
				"exchange", models.ExchangeSlugKucoin.String(),
				"orderID", args.InternalOrder.ID.String(),
			)
			order.FailReason = "Order is older than 7 days, cannot retrieve history from API"
			return order, nil
		}
	}
	if args.ExternalOrderID != nil {
		return o.getOrderDetails(ctx, args.InstrumentID, order, *args.ExternalOrderID, false)
	}
	if args.ClientOrderID != nil {
		return o.getOrderDetails(ctx, args.InstrumentID, order, *args.ClientOrderID, true)
	}
	return order, nil
}

func (o *Service) getOrderDetails(ctx context.Context, symbol *string, order *models.OrderDetailsDTO, orderID string, external bool) (*models.OrderDetailsDTO, error) {
	exchangeOrder := &kucoinmodels.Order{}

	if external { //nolint:nestif
		res, err := o.exClient.Spot().GetOrderByClientOID(ctx, kucoinrequests.GetOrderByClientOID{
			ClientOrderOID: orderID,
			Symbol:         *symbol,
		})
		if err != nil {
			return nil, err
		}
		if res.Order == nil {
			return order, nil
		}
		exchangeOrder = res.Order
	} else {
		res, err := o.exClient.Spot().GetOrderByOrderID(ctx, kucoinrequests.GetOrderByOrderID{
			OrderID: orderID,
			Symbol:  *symbol,
		})
		if err != nil {
			return nil, err
		}
		if res.Order == nil {
			return order, nil
		}
		exchangeOrder = res.Order
	}

	if symbol != nil { //nolint:nestif
		symbolInfo, err := o.exClient.Public().GetSymbol(ctx, kucoinrequests.GetSymbol{
			Symbol: *symbol,
		})
		if err != nil || symbolInfo.Symbol == nil {
			return nil, err
		}
		if strings.Contains(symbolInfo.Symbol.QuoteCurrency, "USD") {
			order.AmountUSD = exchangeOrder.DealFunds
		} else {
			tickerInfo, err := o.exClient.Public().GetTicker(ctx, kucoinrequests.GetTicker{
				Symbol: strings.Join([]string{symbolInfo.Symbol.BaseCurrency, "USDT"}, "-"),
			})
			if err != nil || tickerInfo.Ticker == nil {
				return nil, err
			}
			order.AmountUSD = exchangeOrder.DealSize.Mul(tickerInfo.Ticker.BestAsk)
		}
	}

	order.Amount = exchangeOrder.DealSize

	// TODO: check if order is canceled and etc.
	if exchangeOrder.Active {
		order.State = models.ExchangeOrderStatusInProgress
	} else {
		order.State = models.ExchangeOrderStatusCompleted
	}

	return order, nil
}

func (o *Service) GetWithdrawalRules(ctx context.Context, coins ...string) ([]*models.WithdrawalRulesDTO, error) {
	currEnabled, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugKucoin)
	if err != nil {
		return nil, err
	}

	currEnabled = lo.Filter(currEnabled, func(item *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) bool {
		return lo.Contains(coins, item.ID.String)
	})

	currReferences := make([]*kucoinmodels.Currency, 0, len(coins))
	for _, coin := range currEnabled {
		res, err := o.exClient.Market().GetCurrency(ctx, kucoinrequests.GetCurrency{
			Currency: coin.Ticker,
			Chain:    coin.Chain,
		})
		if err != nil {
			return nil, err
		}
		currReferences = append(currReferences, res.Currency)
	}

	exchangeRules := make([]*models.WithdrawalRulesDTO, 0, len(currReferences))
	for _, curr := range currReferences {
		if slices.ContainsFunc(currEnabled, func(ec *repo_exchange_chains.GetEnabledCurrenciesRow) bool { //nolint:nestif
			return ec.Ticker == curr.Currency
		}) {
			for _, chain := range curr.Chains {
				if slices.ContainsFunc(currEnabled, func(ec *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
					return ec.Chain == chain.ChainID && ec.Ticker == curr.Currency
				}) {
					rule := &models.WithdrawalRulesDTO{
						Currency:           curr.Currency,
						Chain:              chain.ChainID,
						MinDepositAmount:   chain.DepositMinSize.String(),
						MinWithdrawAmount:  chain.WithdrawalMinSize.String(),
						NumOfConfirmations: strconv.Itoa(chain.Confirms),
						WithdrawPrecision:  strconv.Itoa(chain.WithdrawPrecision),
						Fee:                chain.WithdrawalMinFee.String(),
						WithdrawFeeType:    models.WithdrawalFeeTypeFixed,
					}

					if chain.DepositMinSize.IsZero() {
						// If deposit minimum is zero or empty, set it to the 1 USDT equivalent
						minDepositAmount, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
							Source:     models.ExchangeSlugKucoin.String(),
							From:       models.CurrencyCodeUSDT,
							To:         curr.Currency,
							Amount:     "1",
							StableCoin: false,
						})
						if err != nil {
							return nil, fmt.Errorf("convert 1 USDT to %s: %w", curr.Currency, err)
						}
						if minDepositAmount.LessThan(decimal.NewFromFloat(1.1)) && minDepositAmount.GreaterThan(decimal.NewFromFloat(0.9)) {
							// If the converted amount is around 1, set it to 1
							minDepositAmount = decimal.NewFromFloat(1)
						} else {
							minDepositAmount = minDepositAmount.RoundUp(int32(chain.WithdrawPrecision))
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

func (o *Service) GetWithdrawalByID(ctx context.Context, args *models.GetWithdrawalByIDParams) (*models.WithdrawalStatusDTO, error) {
	if args.ClientOrderID == nil {
		return nil, fmt.Errorf("client order ID is required")
	}
	res, err := o.exClient.Account().GetWithdrawalHistory(ctx, kucoinrequests.GetWithdrawalHistory{})
	if err != nil {
		return nil, err
	}
	if len(res.History.Items) == 0 {
		return nil, fmt.Errorf("withdrawal not found")
	}
	for _, withdrawal := range res.History.Items {
		if withdrawal.ID == *args.ClientOrderID {
			return &models.WithdrawalStatusDTO{
				ID:           withdrawal.ID,
				TxHash:       withdrawal.WalletTxID,
				NativeAmount: withdrawal.Amount,
				Status:       withdrawal.Status.String(),
			}, nil
		}
	}
	return nil, fmt.Errorf("withdrawal ID %s not found", *args.ClientOrderID)
}

func (o *Service) GetConnectionHash() string {
	return o.connHash
}
