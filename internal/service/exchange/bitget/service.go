package bitget

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
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget"
	bitgetclients "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/clients"
	bitgetmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/models"
	bitgetrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/requests"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

var (
	ErrInsufficientBalance         = errors.New("insufficient balance")
	ErrUnprocessableCurrencyStatus = errors.New("unprocessable currency status")
	ErrMaxOrderValueReached        = errors.New("max order value reached")
	ErrMinWithdrawalBalance        = errors.New("withdrawal threshold not met")
	ErrWithdrawalBalanceLocked     = errors.New("withdrawal balance locked")
)

const (
	WithdrawalStep = 10
)

func NewService(l logger.Logger, accessKey, secretKey, passPhrase string, public bool, baseURL *url.URL, storage storage.IStorage, store limiter.Store, convSvc currconv.ICurrencyConvertor) (*Service, error) {
	exClient, err := bitgetclients.NewBaseClient(&bitgetclients.ClientOptions{
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		PassPhrase: passPhrase,
		BaseURL:    baseURL,
		Public:     public,
	}, store)
	if err != nil {
		return nil, err
	}

	connHash, err := hash.SHA256ConnectionHash(models.ExchangeSlugBitget.String(), accessKey, secretKey, passPhrase)
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
	exClient bitgetclients.IBaseBitgetClient
	storage  storage.IStorage
	convSvc  currconv.ICurrencyConvertor
	l        logger.Logger
	connHash string
}

func (o *Service) TestConnection(ctx context.Context) error {
	_, err := o.exClient.Spot().Account().AccountAssets(ctx, &bitgetrequests.AccountAssetsRequest{
		AssetType: "hold_only",
	})
	if err != nil {
		return err
	}
	return nil
}

func (o *Service) GetAccountBalance(ctx context.Context) ([]*models.AccountBalanceDTO, error) {
	enabledCurrencies, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugBitget)
	if err != nil {
		return nil, err
	}

	res, err := o.exClient.Spot().Account().AccountAssets(ctx, &bitgetrequests.AccountAssetsRequest{
		AssetType: "hold_only",
	})
	if err != nil {
		return nil, err
	}

	var dto []*models.AccountBalanceDTO
	for _, coin := range res.Data {
		if slices.IndexFunc(enabledCurrencies, func(c *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
			return c.Ticker == coin.Coin
		}) == -1 {
			continue
		}
		if !coin.Available.IsZero() {
			amountUSD, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
				Source:     models.ExchangeSlugBitget.String(),
				From:       coin.Coin,
				To:         "USDT",
				Amount:     coin.Available.String(),
				StableCoin: false,
			})
			if err != nil {
				return nil, err
			}
			currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByTicker(ctx, coin.Coin)
			if err != nil {
				continue
			}
			dto = append(dto, &models.AccountBalanceDTO{
				Currency:  currencyID,
				Type:      models.CurrencyTypeCrypto.String(),
				Amount:    coin.Available,
				AmountUSD: amountUSD.Round(4),
			})
		}
	}

	return dto, nil
}

func (o *Service) GetCurrencyBalance(ctx context.Context, currency string) (*decimal.Decimal, error) {
	res, err := o.exClient.Spot().Account().AccountAssets(ctx, &bitgetrequests.AccountAssetsRequest{
		Coin:      currency,
		AssetType: "all",
	})
	if err != nil {
		return nil, err
	}
	for _, data := range res.Data {
		if strings.EqualFold(data.Coin, currency) {
			return &data.Available, nil
		}
	}

	return &decimal.Zero, nil
}

func (o *Service) GetExchangeSymbols(ctx context.Context) ([]*models.ExchangeSymbolDTO, error) {
	res, err := o.exClient.Spot().Market().SymbolInformation(ctx, &bitgetrequests.SymbolInformationRequest{})
	if err != nil {
		return nil, err
	}

	dto := make([]*models.ExchangeSymbolDTO, 0, len(res.Data)*2)
	for _, s := range res.Data {
		if s.Status != bitgetmodels.SymbolStatusOnline {
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

func (o *Service) GetDepositAddresses(ctx context.Context, currency, _ string) ([]*models.DepositAddressDTO, error) {
	enabledCurrencies, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugBitget)
	if err != nil {
		return nil, fmt.Errorf("fetch enabled currencies: %w", err)
	}
	filteredCurrencies := lo.Filter(enabledCurrencies, func(i *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) bool {
		return i.Ticker == currency
	})

	deposits := make([]*models.DepositAddressDTO, 0, len(filteredCurrencies))
	for _, network := range filteredCurrencies {
		res, err := o.exClient.Spot().Account().DepositAddress(ctx, &bitgetrequests.DepositAddressRequest{
			Coin:  currency,
			Chain: network.Chain,
		})
		if err != nil {
			return nil, err
		}
		currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByParams(ctx, repo_exchange_chains.GetCurrencyIDByParamsParams{
			Ticker: network.Ticker,
			Chain:  network.Chain,
			Slug:   models.ExchangeSlugBitget,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get internal currency id %s: %w", network.Chain, err)
		}
		deposits = append(deposits, &models.DepositAddressDTO{
			Address:          res.Data.Address,
			Currency:         currencyID,
			Chain:            res.Data.Chain,
			AddressType:      models.DepositAddress,
			InternalCurrency: res.Data.Coin,
		})
	}

	return deposits, nil
}

func (o *Service) CreateWithdrawalOrder(ctx context.Context, args *models.CreateWithdrawalOrderParams) (*models.ExchangeWithdrawalDTO, error) {
	args.NativeAmount = args.NativeAmount.RoundDown(int32(args.WithdrawalPrecision)) //nolint:gosec

	internalCurrency, err := o.storage.ExchangeChains().GetTickerByCurrencyID(ctx, repo_exchange_chains.GetTickerByCurrencyIDParams{
		CurrencyID: args.Currency,
		Slug:       models.ExchangeSlugBitget,
	})
	if err != nil {
		return nil, err
	}

	req := &bitgetrequests.WalletWithdrawalRequest{
		Coin:         internalCurrency,
		Chain:        args.Chain,
		Address:      args.Address,
		TransferType: "on_chain",
		Size:         args.NativeAmount.Sub(args.Fee).String(),
	}

	clientOrderID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	req.ClientOID = clientOrderID.String()

	o.l.Info("withdrawal request assembled",
		"exchange", models.ExchangeSlugBitget.String(),
		"recordID", args.RecordID.String(),
		"request", req,
		"amount", req.Size,
		"currency", req.Coin,
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
			return nil, ErrWithdrawalBalanceLocked
		}

		withdrawalStep, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
			Source:     models.ExchangeSlugBitget.String(),
			From:       "USDT",
			To:         req.Coin,
			Amount:     decimal.NewFromInt(WithdrawalStep).String(),
			StableCoin: false,
		})
		if err != nil {
			return nil, err
		}

		req.Size = amount.String()
		res, err := o.exClient.Spot().Account().WalletWithdrawal(ctx, req)
		if err == nil {
			dto.InternalOrderID = res.Data.ClientOid
			dto.ExternalOrderID = res.Data.OrderID
			return dto, nil
		}

		if strings.Contains(err.Error(), "temporarily frozen") {
			o.l.Error("insufficient funds, retrying with reduced amount",
				ErrWithdrawalBalanceLocked,
				"exchange", models.ExchangeSlugBitget.String(),
				"recordID", args.RecordID.String(),
				"current_amount", amount.String(),
			)

			amount = amount.Sub(withdrawalStep)
			if amount.LessThan(minWithdrawal) {
				return nil, ErrMinWithdrawalBalance
			}
			dto.RetryReason = ErrWithdrawalBalanceLocked.Error()
			continue
		}

		return nil, err
	}
}

func (o *Service) CreateSpotOrder(ctx context.Context, from string, to string, side string, ticker string, amount *decimal.Decimal, rule *models.OrderRulesDTO) (*models.ExchangeOrderDTO, error) { //nolint:all
	orderID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	res, err := o.exClient.Spot().Market().SymbolInformation(ctx, &bitgetrequests.SymbolInformationRequest{
		Symbol: ticker,
	})
	if err != nil {
		return nil, err
	}
	if len(res.Data) == 0 {
		return nil, fmt.Errorf("symbol %s not found", ticker)
	}

	symbol := res.Data[0]

	if symbol.Status != bitgetmodels.SymbolStatusOnline {
		return nil, fmt.Errorf("%w: symbol is not online", ErrUnprocessableCurrencyStatus)
	}

	spotOrderRequest := &bitgetrequests.PlaceOrderRequest{
		ClientOID: orderID.String(),
		OrderType: bitgetmodels.OrderTypeMarket,
		Side:      bitgetmodels.OrderSide(side),
		Symbol:    symbol.Symbol,
	}

	baseBalance, quoteBalance := decimal.Zero, decimal.Zero //nolint:all
	{
		res, err := o.exClient.Spot().Account().AccountAssets(ctx, &bitgetrequests.AccountAssetsRequest{
			Coin:      symbol.BaseCoin,
			AssetType: "all",
		})
		if err != nil {
			return nil, err
		}
		baseBalance = res.Data[0].Available
	}
	{
		res, err := o.exClient.Spot().Account().AccountAssets(ctx, &bitgetrequests.AccountAssetsRequest{
			Coin:      symbol.QuoteCoin,
			AssetType: "all",
		})
		if err != nil {
			return nil, err
		}
		quoteBalance = res.Data[0].Available
	}

	orderMinimumBase, err := decimal.NewFromString(rule.MinOrderAmount)
	if err != nil {
		return nil, err
	}

	orderMinimumQuote, err := decimal.NewFromString(rule.MinOrderValue)
	if err != nil {
		return nil, err
	}

	maxAmount := decimal.Zero
	switch spotOrderRequest.Side {
	case bitgetmodels.OrderSideSell:
		maxAmount = baseBalance
		if maxAmount.LessThan(orderMinimumBase) {
			return nil, ErrInsufficientBalance
		}
	case bitgetmodels.OrderSideBuy:
		maxAmount = quoteBalance
		if maxAmount.LessThan(orderMinimumQuote) {
			return nil, ErrInsufficientBalance
		}
	default:
		return nil, fmt.Errorf("unsupported order type %s", spotOrderRequest.Side)
	}

	amount = &maxAmount //nolint:staticcheck

	switch spotOrderRequest.Side {
	case bitgetmodels.OrderSideSell:
		spotOrderRequest.Size = maxAmount.RoundDown(int32(rule.AmountPrecision)).String() //nolint:gosec
	case bitgetmodels.OrderSideBuy:
		spotOrderRequest.Size = maxAmount.RoundDown(int32(rule.ValuePrecision)).String() //nolint:gosec
	}

	placedOrder, err := o.exClient.Spot().Trade().PlaceOrder(ctx, spotOrderRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	order := &models.ExchangeOrderDTO{
		ExchangeOrderID: placedOrder.Data.OrderID,
		ClientOrderID:   orderID.String(),
		Amount:          *amount,
	}

	return order, nil
}

func (o *Service) GetOrderRule(ctx context.Context, symbol string) (*models.OrderRulesDTO, error) {
	symbolData, err := o.exClient.Spot().Market().SymbolInformation(ctx, &bitgetrequests.SymbolInformationRequest{
		Symbol: symbol,
	})
	if err != nil {
		return nil, err
	}
	if len(symbolData.Data) == 0 {
		return nil, fmt.Errorf("symbol %s not found", symbol)
	}
	s := symbolData.Data[0]

	// Fix for rounded minTradeAmount being 0 all the time
	{
		if s.QuoteCoin != "USDT" { //nolint:nestif
			base, err := o.exClient.Spot().Market().TickerInformation(ctx, &bitgetrequests.TickerInformationRequest{
				Symbol: s.BaseCoin + "USDT",
			})
			if err != nil {
				return nil, err
			}
			if len(base.Data) == 0 {
				return nil, fmt.Errorf("ticker %s not found", symbol)
			}

			quote, err := o.exClient.Spot().Market().TickerInformation(ctx, &bitgetrequests.TickerInformationRequest{
				Symbol: s.QuoteCoin + "USDT",
			})
			if err != nil {
				return nil, err
			}
			if len(quote.Data) == 0 {
				return nil, fmt.Errorf("ticker %s not found", symbol)
			}

			minUSDT := s.MinTradeUSDT.Copy().Mul(decimal.NewFromFloat(1.10))
			s.MinTradeUSDT = minUSDT.Copy().Div(quote.Data[0].AskPr)
			s.MinTradeAmount = minUSDT.Copy().Div(base.Data[0].AskPr)
		} else {
			res, err := o.exClient.Spot().Market().TickerInformation(ctx, &bitgetrequests.TickerInformationRequest{
				Symbol: symbol,
			})
			if err != nil {
				return nil, err
			}
			if len(res.Data) == 0 {
				return nil, fmt.Errorf("ticker %s not found", symbol)
			}

			s.MinTradeUSDT = s.MinTradeUSDT.Mul(decimal.NewFromFloat(1.10))
			s.MinTradeAmount = s.MinTradeUSDT.Div(res.Data[0].AskPr)
		}
	}

	return &models.OrderRulesDTO{
		Symbol:          s.Symbol,
		State:           s.Status.String(),
		BaseCurrency:    s.BaseCoin,
		QuoteCurrency:   s.QuoteCoin,
		PricePrecision:  s.PricePrecision,
		AmountPrecision: s.QuantityPrecision,
		ValuePrecision:  s.QuotePrecision,
		MinOrderAmount:  s.MinTradeAmount.String(),
		MaxOrderAmount:  s.MaxTradeAmount.String(),
		MinOrderValue:   s.MinTradeUSDT.String(),
	}, nil
}

func (o *Service) GetOrderDetails(ctx context.Context, args *models.GetOrderByIDParams) (*models.OrderDetailsDTO, error) {
	order := &models.OrderDetailsDTO{
		State:     models.ExchangeOrderStatusFailed,
		Amount:    decimal.Zero,
		AmountUSD: decimal.Zero,
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
	req := &bitgetrequests.OrderInformationRequest{}
	if external {
		req.ClientOID = orderID
	} else {
		req.OrderID = orderID
	}
	res, err := o.exClient.Spot().Trade().OrderInformation(ctx, req)
	if err != nil && errors.Is(err, bitget.ErrParameterOrderID) || errors.Is(err, bitget.ErrParameterClientOrderID) {
		return order, nil
	}
	if err != nil {
		return nil, err
	}
	if len(res.Data) == 0 {
		return order, nil
	}

	orderData := res.Data[0]

	if symbol != nil { //nolint:nestif
		symbolInfo, err := o.exClient.Spot().Market().SymbolInformation(ctx, &bitgetrequests.SymbolInformationRequest{
			Symbol: *symbol,
		})
		if err != nil || len(symbolInfo.Data) == 0 {
			return nil, err
		}
		symbolData := symbolInfo.Data[0]
		if strings.Contains(symbolData.QuoteCoin, "USD") { // slightly better then "USDT/USDC"
			order.AmountUSD = orderData.QuoteVolume
		} else {
			tickerInfo, err := o.exClient.Spot().Market().TickerInformation(ctx, &bitgetrequests.TickerInformationRequest{
				Symbol: strings.Join([]string{symbolData.BaseCoin, "USDT"}, ""),
			})
			if err != nil || len(tickerInfo.Data) == 0 {
				return nil, err
			}
			tickerData := tickerInfo.Data[0]
			order.AmountUSD = orderData.BaseVolume.Mul(tickerData.AskPr)
		}
	}

	order.Amount = orderData.BaseVolume

	switch res.Data[0].Status {
	case bitgetmodels.OrderStatusFilled:
		order.State = models.ExchangeOrderStatusCompleted
	case bitgetmodels.OrderStatusLive:
		order.State = models.ExchangeOrderStatusInProgress
	case bitgetmodels.OrderStatusCanceled:
		order.State = models.ExchangeOrderStatusFailed
	case bitgetmodels.OrderStatusPartiallyFilled:
		order.State = models.ExchangeOrderStatusInProgress
	default:
		order.State = models.ExchangeOrderStatusInProgress
	}
	return order, nil
}

func (o *Service) GetWithdrawalRules(ctx context.Context, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	currEnabled, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugBitget)
	if err != nil {
		return nil, err
	}

	currEnabled = lo.Filter(currEnabled, func(i *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) bool {
		return lo.Contains(currencies, i.ID.String)
	})

	ccys := lo.Map(currEnabled, func(i *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) string {
		return i.Ticker
	})

	currReferences := make([]*bitgetmodels.CoinInformation, 0, len(ccys))
	for _, currency := range ccys {
		res, err := o.exClient.Spot().Market().CoinInformation(ctx, &bitgetrequests.CoinInformationRequest{
			Coin: currency,
		})
		if err != nil {
			return nil, err
		}
		currReferences = append(currReferences, res.Data...)
	}

	exchangeRules := make([]*models.WithdrawalRulesDTO, 0, len(currReferences))
	for _, item := range currReferences {
		if slices.ContainsFunc(currEnabled, func(c *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
			return c.Ticker == item.Coin
		}) {
			for _, network := range item.Chains {
				if slices.ContainsFunc(currEnabled, func(c *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
					return c.Chain == network.Chain && item.Coin == c.Ticker
				}) {
					exchangeRules = append(exchangeRules, &models.WithdrawalRulesDTO{
						Currency:           item.Coin,
						Chain:              network.Chain,
						MinDepositAmount:   network.MinDepositAmount.String(),
						MinWithdrawAmount:  network.MinWithdrawAmount.String(),
						NumOfConfirmations: network.WithdrawConfirm.String(),
						WithdrawPrecision:  network.WithdrawMinScale.String(),
						Fee:                network.WithdrawFee.String(),
						WithdrawFeeType:    models.WithdrawalFeeTypeFixed,
					})
				}
			}
		}
	}

	return exchangeRules, nil
}

func (o *Service) GetWithdrawalByID(ctx context.Context, args *models.GetWithdrawalByIDParams) (*models.WithdrawalStatusDTO, error) {
	request := &bitgetrequests.WithdrawalRecordsRequest{
		StartTime: strconv.FormatInt(time.Now().Add(-time.Hour*48).UnixMilli(), 10),
		EndTime:   strconv.FormatInt(time.Now().UnixMilli(), 10),
	}
	if args.ClientOrderID == nil {
		return nil, fmt.Errorf("ClientOrderID is required")
	}
	if args.ClientOrderID != nil {
		request.ClientOid = *args.ClientOrderID
	}

	wdRecords, err := o.exClient.Spot().Account().WithdrawalRecords(ctx, request)
	if err != nil {
		return nil, err
	}
	if len(wdRecords.Data) == 0 {
		return nil, fmt.Errorf("withdrawal not found")
	}
	wdRecord := wdRecords.Data[0]

	res := &models.WithdrawalStatusDTO{
		ID:     wdRecord.OrderID,
		Status: wdRecord.Status.String(),
	}
	if wdRecord.TradeID != "" {
		res.TxHash = wdRecord.TradeID
	}
	if !wdRecord.Size.IsZero() {
		res.NativeAmount = wdRecord.Size
	}
	return res, nil
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
