package okx

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
	"github.com/dv-net/dv-merchant/pkg/exchange_client/okx"
	okxmodels "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/models"
	okxrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/okx/requests"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/ulule/limiter/v3"
)

const (
	WithdrawalStep = 10
)

type Service struct {
	exClient *okx.BaseClient
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
	if args.InstrumentID == nil {
		return order, nil
	}
	req := okxrequests.OrderDetails{
		InstID: *args.InstrumentID,
	}
	if args.ExternalOrderID != nil {
		req.OrdID = *args.ExternalOrderID
	}
	if args.ClientOrderID != nil {
		req.ClOrdID = *args.ClientOrderID
	}
	res, err := o.exClient.Trade().GetOrderDetail(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(res.Orders) > 0 {
		orderData := res.Orders[0]
		amt, err := decimal.NewFromString(orderData.Sz)
		if err != nil {
			return nil, err
		}
		order.Amount = amt

		amtUSD, err := o.getNotionalUSD(ctx, orderData.InstID, amt)
		if err != nil {
			return nil, err
		}
		order.AmountUSD = amtUSD

		switch orderData.State {
		case okxmodels.OrderStateFilled:
			order.State = models.ExchangeOrderStatusCompleted
		default:
			order.State = models.ExchangeOrderStatusInProgress
		}
	}
	return order, nil
}

func (o *Service) getNotionalUSD(ctx context.Context, instID string, amount decimal.Decimal) (decimal.Decimal, error) {
	res, err := o.exClient.Market().GetIndexTickers(ctx, okxrequests.GetIndexTickers{
		QuoteCcy: "USDT",
	})
	if err != nil {
		return decimal.Zero, err
	}

	curr, exists := lo.Find(res.IndexTickers, func(item *okxmodels.IndexTicker) bool {
		return strings.Split(instID, "-")[0]+"-"+"USDT" == item.InstID
	})
	if !exists {
		return decimal.Zero, errors.New("currency not found")
	}

	amt, err := decimal.NewFromString(curr.IdxPx)
	if err != nil {
		return decimal.Zero, err
	}
	return amount.Mul(amt), nil
}

func (o *Service) GetOrderRule(ctx context.Context, symbol string) (*models.OrderRulesDTO, error) {
	res, err := o.exClient.Public().GetInstruments(ctx, okxrequests.GetInstruments{
		InstID:   symbol,
		InstType: "SPOT",
	})
	if err != nil {
		return nil, err
	}
	if len(res.Instruments) == 0 {
		return nil, errors.New("symbol not found")
	}
	symbolData := res.Instruments[0]
	dto := &models.OrderRulesDTO{
		Symbol:        symbolData.InstID,
		State:         symbolData.State,
		BaseCurrency:  symbolData.BaseCcy,
		QuoteCurrency: symbolData.QuoteCcy,
	}
	if v, err := decimal.NewFromString(symbolData.MinSz); err == nil {
		dto.MinOrderAmount = v.String()
	}
	if v, err := decimal.NewFromString(symbolData.MaxMktSz); err == nil {
		dto.MaxOrderAmount = v.String()
	}

	orderMinimumQuote := decimal.Zero //nolint:staticcheck
	orderMinimumBase, err := decimal.NewFromString(symbolData.MinSz)
	if err != nil {
		return nil, fmt.Errorf("failed to parse minimum order size: %w", err)
	}

	r, err := o.exClient.Market().GetIndexTickers(ctx, okxrequests.GetIndexTickers{
		InstID: symbolData.InstID,
	})
	if err != nil || len(r.IndexTickers) == 0 {
		return nil, fmt.Errorf("failed to get index tickers: %w", err)
	}

	qbRate, err := decimal.NewFromString(r.IndexTickers[0].IdxPx)
	if err != nil {
		return nil, err
	}
	orderMinimumQuote = orderMinimumBase.Mul(qbRate)

	dto.MinOrderValue = orderMinimumQuote.String()
	return dto, nil
}

func (o *Service) GetWithdrawalByID(ctx context.Context, args *models.GetWithdrawalByIDParams) (*models.WithdrawalStatusDTO, error) {
	request := &okxrequests.GetWithdrawalHistory{}
	if args.ClientOrderID == nil && args.ExternalOrderID == nil {
		return nil, errors.New("ClientOrderID or ExternalOrderID is required")
	}
	if args.ClientOrderID != nil {
		request.ClientOrderID = *args.ClientOrderID
	}
	if args.ExternalOrderID != nil {
		request.WithdrawalID = *args.ExternalOrderID
	}

	wdRecord, err := o.exClient.Funding().GetWithdrawalHistory(ctx, request)
	if err != nil {
		return nil, err
	}
	if len(wdRecord.WithdrawalHistories) < 1 {
		return nil, errors.New("withdrawal not found")
	}
	singleWdRecord := wdRecord.WithdrawalHistories[0]
	res := &models.WithdrawalStatusDTO{
		ID:     strconv.FormatInt(singleWdRecord.WdID, 10),
		Status: singleWdRecord.State.String(),
	}
	if singleWdRecord.TxID != "" {
		res.TxHash = singleWdRecord.TxID
	}
	return res, nil
}

func (o *Service) GetCurrencyBalance(ctx context.Context, currency string) (*decimal.Decimal, error) {
	assetBalances, err := o.exClient.Account().GetBalance(ctx, okxrequests.GetAccountBalance{
		Ccy: []string{currency},
	})
	if err != nil {
		return nil, err
	}
	fundingBalances, err := o.exClient.Funding().GetBalance(ctx, okxrequests.GetFundingBalance{
		Ccy: []string{currency},
	})
	if err != nil {
		return nil, err
	}

	assetAmt, fundingAmt := decimal.Zero, decimal.Zero
	if len(fundingBalances.Balances) > 0 {
		fundingAmt, err = decimal.NewFromString(fundingBalances.Balances[0].AvailBal)
		if err != nil {
			return nil, err
		}
	}

	if len(assetBalances.Balances) > 0 {
		if len(assetBalances.Balances[0].Details) > 0 {
			assetAmt, err = decimal.NewFromString(assetBalances.Balances[0].Details[0].AvailBal)
			if err != nil {
				return nil, err
			}
		}
	}
	amt := assetAmt.Add(fundingAmt)
	return &amt, nil
}

func (o *Service) GetWithdrawalRules(ctx context.Context, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	currEnabled, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugOkx)
	if err != nil {
		return nil, err
	}

	currEnabled = lo.Filter(currEnabled, func(i *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) bool {
		return lo.Contains(currencies, i.ID.String)
	})

	ccys := lo.Map(currEnabled, func(i *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) string {
		return i.Ticker
	})

	currReferences, err := o.exClient.Funding().GetCurrencies(ctx, okxrequests.GetCurrencies{
		Ccy: slices.Compact(ccys),
	})
	if err != nil {
		return nil, err
	}

	exchangeRules := lo.FilterMap(currReferences.Currencies, func(item *okxmodels.Currency, _ int) (*models.WithdrawalRulesDTO, bool) {
		if lo.ContainsBy(currEnabled, func(c *repo_exchange_chains.GetEnabledCurrenciesRow) bool {
			return c.Chain == item.Chain && c.Ticker == item.Ccy
		}) {
			return &models.WithdrawalRulesDTO{
				Currency:           item.Ccy,
				Chain:              item.Chain,
				MinDepositAmount:   item.MinDep,
				MinWithdrawAmount:  item.MinWd,
				NumOfConfirmations: item.MinDepArrivalConfirm,
				WithdrawPrecision:  item.WdTckSz,
				Fee:                item.Fee,
			}, true
		}
		return nil, false
	})
	return exchangeRules, nil
}

func (o *Service) CreateSpotOrder(ctx context.Context, baseSymbol string, quoteSymbol string, side string, ticker string, amount *decimal.Decimal, _ *models.OrderRulesDTO) (*models.ExchangeOrderDTO, error) { //nolint:staticcheck
	orderID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}
	clientOrderID := strings.ReplaceAll(orderID.String(), "-", "")

	instrumentRequest := okxrequests.GetInstruments{
		InstType: "SPOT",
		InstID:   ticker,
	}

	instrumentsData, err := o.exClient.Public().GetInstruments(ctx, instrumentRequest)
	if err != nil {
		// instrumentRequest.InstID = quoteSymbol + "-" + baseSymbol
		// instrumentsData, err = o.exClient.Public().GetInstruments(ctx, instrumentRequest)
		return nil, err
	}
	if len(instrumentsData.Instruments) == 0 {
		return nil, fmt.Errorf("instrument %s not found", instrumentRequest.InstID)
	}

	symbolData := instrumentsData.Instruments[0]
	if symbolData.State != "live" {
		return nil, fmt.Errorf("instrument %s is not live: %s", baseSymbol+"-"+quoteSymbol, symbolData.State)
	}

	spotOrderRequest := okxrequests.PlaceOrder{
		InstID:  symbolData.InstID,
		ClOrdID: clientOrderID,
		TdMode:  okxmodels.CashTradingMode.String(),
		Side:    side,
		OrdType: okxmodels.OrderTypeMarket.String(),
	}

	orderMinimumQuote := decimal.Zero
	orderMinimumBase, err := decimal.NewFromString(symbolData.MinSz)
	if err != nil {
		return nil, fmt.Errorf("failed to parse minimum order size: %w", err)
	}

	r, err := o.exClient.Market().GetIndexTickers(ctx, okxrequests.GetIndexTickers{
		InstID: symbolData.InstID,
	})
	if err != nil || len(r.IndexTickers) == 0 {
		return nil, fmt.Errorf("failed to get index tickers: %w", err)
	}

	qbRate, err := decimal.NewFromString(r.IndexTickers[0].IdxPx)
	if err != nil {
		return nil, err
	}
	bqRate := decimal.NewFromInt(1).Div(qbRate)

	switch spotOrderRequest.Side {
	case okxmodels.OrderSideBuy.String():
		orderMinimumQuote = orderMinimumBase
	case okxmodels.OrderSideSell.String():
		orderMinimumQuote = orderMinimumBase.Mul(bqRate)
	}

	ccys := []string{symbolData.BaseCcy, symbolData.QuoteCcy}
	spotBalances, err := o.exClient.Account().GetBalance(ctx, okxrequests.GetAccountBalance{
		Ccy: ccys,
	})
	if err != nil {
		return nil, err
	}

	fundingBalances, err := o.exClient.Funding().GetBalance(ctx, okxrequests.GetFundingBalance{
		Ccy: ccys,
	})
	if err != nil {
		return nil, err
	}

	fundingFiltered := lo.FilterMap(fundingBalances.Balances, func(item *okxmodels.FundingBalance, _ int) (*okxmodels.FundingBalance, bool) {
		return item, lo.Contains(ccys, item.Ccy)
	})

	spotFiltered := lo.FilterMap(spotBalances.Balances[0].Details, func(item *okxmodels.BalanceDetails, _ int) (*okxmodels.BalanceDetails, bool) {
		return item, lo.Contains(ccys, item.Ccy)
	})

	fundsTransferRequest := &okxrequests.FundsTransfer{
		From: okxmodels.BeneficiaryAccountTypeFunding.Int(),
		To:   okxmodels.BeneficiaryAccountTypeTrading.Int(),
	}

	maxAmount := decimal.Zero     //nolint:staticcheck
	spotAmount := decimal.Zero    //nolint:staticcheck
	fundingAmount := decimal.Zero //nolint:staticcheck

	switch spotOrderRequest.Side {
	case okxmodels.OrderSideBuy.String():
		fundsTransferRequest.Ccy = symbolData.QuoteCcy
		spotAmount = o.getSpotBalance(symbolData.QuoteCcy, spotFiltered)
		fundingAmount = o.getFundingBalance(symbolData.QuoteCcy, fundingFiltered)
		maxAmount = fundingAmount.Add(spotAmount)
		if maxAmount.LessThan(orderMinimumQuote) {
			return nil, exchangeclient.ErrMinWithdrawalBalance
		}
	case okxmodels.OrderSideSell.String():
		fundsTransferRequest.Ccy = symbolData.BaseCcy
		spotAmount = o.getSpotBalance(symbolData.BaseCcy, spotFiltered)
		fundingAmount = o.getFundingBalance(symbolData.BaseCcy, fundingFiltered)
		maxAmount = spotAmount.Add(fundingAmount)
		if maxAmount.LessThan(orderMinimumBase) {
			return nil, exchangeclient.ErrMinWithdrawalBalance
		}
	default:
		return nil, fmt.Errorf("unsupported order type %s", spotOrderRequest.Side)
	}

	remainingTopup := maxAmount.Sub(spotAmount)
	if remainingTopup.GreaterThan(decimal.Zero) {
		remainder := remainingTopup.String()
		fundsTransferRequest.Amt = remainder

		_, err = o.exClient.Funding().FundsTransfer(ctx, *fundsTransferRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to transfer funds: %w", err)
		}
	}

	orderSize, _ := maxAmount.Float64()
	spotOrderRequest.Sz = orderSize
	order, err := o.exClient.Trade().PlaceOrder(ctx, []okxrequests.PlaceOrder{spotOrderRequest})
	if err != nil {
		return nil, fmt.Errorf("failed to place order: %w", err)
	}

	amount = &maxAmount //nolint:staticcheck

	return &models.ExchangeOrderDTO{
		ClientOrderID:   order.PlaceOrders[0].ClientOrderID,
		ExchangeOrderID: order.PlaceOrders[0].SystemOrderID,
		Amount:          *amount,
	}, nil
}

func (o *Service) getSpotBalance(symbol string, spot []*okxmodels.BalanceDetails) decimal.Decimal {
	spotAmount := decimal.Zero
	spotBalance, exists := lo.Find(spot, func(item *okxmodels.BalanceDetails) bool {
		return strings.EqualFold(item.Ccy, symbol)
	})
	if exists {
		amount, err := decimal.NewFromString(spotBalance.AvailBal)
		if err != nil {
			return decimal.Zero
		}
		spotAmount = spotAmount.Add(amount)
	}
	return spotAmount
}

func (o *Service) getFundingBalance(symbol string, funding []*okxmodels.FundingBalance) decimal.Decimal {
	fundingAmount := decimal.Zero
	fundingBalance, exists := lo.Find(funding, func(item *okxmodels.FundingBalance) bool {
		return strings.EqualFold(item.Ccy, symbol)
	})
	if exists {
		amount, err := decimal.NewFromString(fundingBalance.AvailBal)
		if err != nil {
			return decimal.Zero
		}
		fundingAmount = fundingAmount.Add(amount)
	}
	return fundingAmount
}

func NewService(l logger.Logger, apiKey, secretKey, passphrase string, baseURL *url.URL, storage storage.IStorage, store limiter.Store, convSvc currconv.ICurrencyConvertor) (*Service, error) {
	exClient := okx.NewBaseClient(&okx.ClientOptions{
		APIKey:     apiKey,
		SecretKey:  secretKey,
		Passphrase: passphrase,
		BaseURL:    baseURL,
	}, store)

	connHash, err := hash.SHA256ConnectionHash(models.ExchangeSlugOkx.String(), apiKey, secretKey, passphrase)
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

func (o *Service) TestConnection(ctx context.Context) error {
	if _, err := o.exClient.Account().GetAccountAndRisks(ctx, okxrequests.GetAccountAndPositionRisk{}); err != nil {
		return err
	}
	return nil
}

type UniversalBalanceDTO struct {
	Balance decimal.Decimal `json:"balance"`
	Ccy     string          `json:"ccy"`
}

func (o *Service) GetAccountBalance(ctx context.Context) ([]*models.AccountBalanceDTO, error) {
	assetBalances, err := o.exClient.Account().GetBalance(ctx, okxrequests.GetAccountBalance{})
	if err != nil {
		return nil, err
	}
	fundingBalances, err := o.exClient.Funding().GetBalance(ctx, okxrequests.GetFundingBalance{})
	if err != nil {
		return nil, err
	}

	balanceMap := make(map[string]*UniversalBalanceDTO)
	for _, balance := range assetBalances.Balances[0].Details {
		amount, err := decimal.NewFromString(balance.AvailBal)
		if err != nil {
			return nil, err
		}
		if record, exists := balanceMap[balance.Ccy]; exists {
			record.Balance = record.Balance.Add(amount)
		} else {
			balanceMap[balance.Ccy] = &UniversalBalanceDTO{
				Ccy:     balance.Ccy,
				Balance: amount,
			}
		}
	}

	for _, balance := range fundingBalances.Balances {
		amount, err := decimal.NewFromString(balance.AvailBal)
		if err != nil {
			return nil, err
		}
		if record, exists := balanceMap[balance.Ccy]; exists {
			record.Balance = record.Balance.Add(amount)
		} else {
			balanceMap[balance.Ccy] = &UniversalBalanceDTO{
				Ccy:     balance.Ccy,
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
			Source:     "okx",
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

func (o *Service) GetExchangeSymbols(ctx context.Context) ([]*models.ExchangeSymbolDTO, error) {
	res, err := o.exClient.Market().GetTickers(ctx, okxrequests.GetTickers{
		InstType: "SPOT", // TODO: hardcoded
	})
	if err != nil {
		return nil, err
	}
	symbols := make([]*models.ExchangeSymbolDTO, 0, len(res.Tickers)*2)
	for _, ticker := range res.Tickers {
		pair := strings.Split(ticker.InstID, "-")
		if len(pair) != 2 {
			continue
		}
		base, quote := pair[0], pair[1]

		symbols = append(symbols, &models.ExchangeSymbolDTO{
			Symbol:      ticker.InstID,
			DisplayName: base + "/" + quote,
			BaseSymbol:  base,
			QuoteSymbol: quote,
			Type:        "sell",
		}, &models.ExchangeSymbolDTO{
			Symbol:      ticker.InstID,
			DisplayName: quote + "/" + base,
			BaseSymbol:  base,
			QuoteSymbol: quote,
			Type:        "buy",
		})
	}
	return symbols, nil
}

func (o *Service) GetDepositAddresses(ctx context.Context, currency, _ string) ([]*models.DepositAddressDTO, error) {
	exchangeAddresses, err := o.exClient.Funding().GetDepositAddress(ctx, okxrequests.GetDepositAddress{
		Ccy: currency,
	})
	if err != nil || len(exchangeAddresses.DepositAddresses) == 0 {
		return nil, err
	}

	addresses := make([]*models.DepositAddressDTO, 0, len(exchangeAddresses.DepositAddresses))
	for _, address := range exchangeAddresses.DepositAddresses {
		currencyID, err := o.storage.ExchangeChains().GetCurrencyIDBySlugAndChain(ctx, repo_exchange_chains.GetCurrencyIDBySlugAndChainParams{
			Chain: address.Chain,
			Slug:  models.ExchangeSlugOkx,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get internal currency id %s: %w", address.Chain, err)
		}
		addresses = append(addresses, &models.DepositAddressDTO{
			Address:          address.Addr,
			Currency:         currencyID,
			Chain:            address.Chain,
			AddressType:      models.DepositAddress,
			InternalCurrency: address.Ccy,
		})
	}
	return addresses, nil
}

func (o *Service) CreateWithdrawalOrder(ctx context.Context, args *models.CreateWithdrawalOrderParams) (*models.ExchangeWithdrawalDTO, error) {
	precision := int32(args.WithdrawalPrecision) //nolint:gosec

	args.NativeAmount = args.NativeAmount.RoundDown(precision).Sub(args.NativeAmount.Div(decimal.NewFromInt(100)).Mul(decimal.NewFromInt(1))).RoundDown(precision)

	internalCurrencyID, err := o.storage.ExchangeChains().GetTickerByCurrencyID(ctx, repo_exchange_chains.GetTickerByCurrencyIDParams{
		CurrencyID: args.Currency,
		Slug:       models.ExchangeSlugOkx,
	})
	if err != nil {
		return nil, err
	}

	funding, err := o.exClient.Funding().GetBalance(ctx, okxrequests.GetFundingBalance{
		Ccy: []string{internalCurrencyID},
	})
	if err != nil {
		return nil, err
	}

	spot, err := o.exClient.Account().GetBalance(ctx, okxrequests.GetAccountBalance{
		Ccy: []string{internalCurrencyID},
	})
	if err != nil {
		return nil, err
	}

	spotBalance := o.getSpotBalance(internalCurrencyID, spot.Balances[0].Details).RoundDown(precision)
	fundingBalance := o.getFundingBalance(internalCurrencyID, funding.Balances).RoundDown(precision)

	totalBalance := spotBalance.Add(fundingBalance)

	o.l.Info("balances",
		"exchange", models.ExchangeSlugOkx.String(),
		"recordID", args.RecordID.String(),
		"withdrawalAmount", args.NativeAmount.String(),
		"withdrawalFee", args.Fee.String(),
		"totalBalance", totalBalance.String(),
		"spotBalance", spotBalance.String(),
		"fundingBalance", fundingBalance.String(),
	)

	if fundingBalance.LessThan(args.NativeAmount) {
		o.l.Info("funding balance is less than withdrawal amount",
			"exchange", models.ExchangeSlugOkx.String(),
			"recordID", args.RecordID.String(),
		)

		if totalBalance.LessThan(args.NativeAmount) {
			o.l.Info("total balance is less than withdrawal amount",
				"exchange", models.ExchangeSlugOkx.String(),
				"recordID", args.RecordID.String(),
			)
			return nil, exchangeclient.ErrMinWithdrawalBalance
		}

		transferAmount := totalBalance.Sub(fundingBalance).RoundDown(precision)
		_, err = o.exClient.Funding().FundsTransfer(ctx, okxrequests.FundsTransfer{
			Ccy:  internalCurrencyID,
			Amt:  transferAmount.String(),
			From: okxmodels.BeneficiaryAccountTypeTrading.Int(),
			To:   okxmodels.BeneficiaryAccountTypeFunding.Int(),
		})

		if err != nil {
			return nil, err
		}
	}

	req := &okxrequests.Withdrawal{
		Ccy:        internalCurrencyID,
		Chain:      args.Chain,
		ToAddr:     args.Address,
		Amt:        args.NativeAmount.Sub(args.Fee).RoundDown(precision).String(),
		WalletType: okxmodels.WalletTypePrivate.String(),
		Dest:       okxmodels.DestinationOnChain.Int(),
	}

	clientOrderID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	req.ClientID = strings.ReplaceAll(clientOrderID.String(), "-", "")

	o.l.Info("withdrawal request assembled",
		"exchange", models.ExchangeSlugOkx.String(),
		"recordID", args.RecordID.String(),
		"request", req,
		"amount", req.Amt,
		"currency", req.Ccy,
		"chain", req.Chain,
		"address", req.ToAddr,
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
			Source:     models.ExchangeSlugOkx.String(),
			From:       "USDT",
			To:         req.Ccy,
			Amount:     decimal.NewFromInt(WithdrawalStep).String(),
			StableCoin: false,
		})
		if err != nil {
			return nil, err
		}

		req.Amt = amount.String()
		order, err := o.exClient.Funding().Withdrawal(ctx, *req)
		if err == nil {
			dto.InternalOrderID = req.ClientID
			dto.ExternalOrderID = order.Withdrawals[0].WdID
			return dto, nil
		}

		if errors.Is(err, exchangeclient.ErrWithdrawalBalanceLocked) {
			o.l.Error("insufficient funds, retrying with reduced amount",
				exchangeclient.ErrWithdrawalBalanceLocked,
				"exchange", models.ExchangeSlugOkx.String(),
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
