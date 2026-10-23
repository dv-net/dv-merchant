package kraken

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	kraken "github.com/dv-net/dv-merchant/pkg/exchange_client/kraken"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/responses"
	"github.com/dv-net/dv-merchant/pkg/iso"
)

// Market buy volume is set in base currency, so part of the quote balance
// is kept aside for price movement between the ticker request and execution.
var buySlippageBuffer = decimal.RequireFromString("0.01")

type pairInfo struct {
	pair  responses.AssetPair
	base  string
	quote string
}

func (o *Service) GetExchangeSymbols(ctx context.Context) ([]*models.ExchangeSymbolDTO, error) {
	pairs, err := o.exClient.Market().GetAssetPairs(ctx)
	if err != nil {
		return nil, fmt.Errorf("get asset pairs: %w", err)
	}

	assets, err := o.assets.get(ctx)
	if err != nil {
		return nil, err
	}

	symbols := make([]*models.ExchangeSymbolDTO, 0, len(pairs)*2)
	for name, pair := range pairs {
		if isEarnAsset(name) || !pair.IsTradable() {
			continue
		}
		base, quote := assets.code(pair.Base), assets.code(pair.Quote)
		if iso.IsFiat(base) || iso.IsFiat(quote) {
			continue
		}

		symbols = append(symbols, &models.ExchangeSymbolDTO{
			Symbol:      pair.AltName,
			DisplayName: base + "/" + quote,
			BaseSymbol:  base,
			QuoteSymbol: quote,
			Type:        models.OrderSideSell.String(),
		}, &models.ExchangeSymbolDTO{
			Symbol:      pair.AltName,
			DisplayName: quote + "/" + base,
			BaseSymbol:  base,
			QuoteSymbol: quote,
			Type:        models.OrderSideBuy.String(),
		})
	}

	return symbols, nil
}

func (o *Service) pairInfo(ctx context.Context, ticker string) (*pairInfo, error) {
	pairs, err := o.exClient.Market().GetAssetPairs(ctx, ticker)
	if err != nil {
		if errors.Is(err, exchangeclient.ErrRateLimited) {
			return nil, exchangeclient.ErrSkipOrder
		}
		return nil, fmt.Errorf("get asset pair %s: %w", ticker, err)
	}

	pair, ok := findPair(pairs, ticker)
	if !ok {
		return nil, fmt.Errorf("symbol %s not found", ticker)
	}

	assets, err := o.assets.get(ctx)
	if err != nil {
		return nil, err
	}

	return &pairInfo{pair: pair, base: assets.code(pair.Base), quote: assets.code(pair.Quote)}, nil
}

// findPair looks the pair up by alt name, because Kraken keys the response by its canonical name.
func findPair(pairs responses.AssetPairsResponse, ticker string) (responses.AssetPair, bool) {
	if pair, ok := pairs[ticker]; ok {
		return pair, true
	}
	for _, pair := range pairs {
		if pair.AltName == ticker || pair.WsName == ticker {
			return pair, true
		}
	}
	return responses.AssetPair{}, false
}

func (o *Service) GetOrderRule(ctx context.Context, ticker string) (*models.OrderRulesDTO, error) {
	info, err := o.pairInfo(ctx, ticker)
	if err != nil {
		return nil, err
	}
	if !info.pair.IsTradable() {
		return nil, exchangeclient.ErrSymbolTradingHalted
	}

	return &models.OrderRulesDTO{
		Symbol:          info.pair.AltName,
		State:           info.pair.Status,
		BaseCurrency:    info.base,
		QuoteCurrency:   info.quote,
		MinOrderAmount:  info.pair.OrderMin.String(),
		MinOrderValue:   info.pair.CostMin.String(),
		AmountPrecision: info.pair.LotDecimals,
		PricePrecision:  info.pair.PairDecimals,
		ValuePrecision:  info.pair.CostDecimals,
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
	if rule == nil {
		var err error
		if rule, err = o.GetOrderRule(ctx, ticker); err != nil {
			return nil, err
		}
	}

	minVolume, err := decimal.NewFromString(rule.MinOrderAmount)
	if err != nil {
		return nil, fmt.Errorf("parse min order amount: %w", err)
	}
	minCost, err := decimal.NewFromString(rule.MinOrderValue)
	if err != nil {
		return nil, fmt.Errorf("parse min order value: %w", err)
	}

	tickers, err := o.exClient.Market().GetTicker(ctx, ticker)
	if err != nil {
		if errors.Is(err, exchangeclient.ErrRateLimited) {
			return nil, exchangeclient.ErrSkipOrder
		}
		return nil, fmt.Errorf("get ticker %s: %w", ticker, err)
	}
	tick, ok := findTicker(tickers, ticker)
	if !ok {
		return nil, fmt.Errorf("ticker %s not found", ticker)
	}

	req := &requests.AddOrderRequest{
		Pair:      ticker,
		Type:      strings.ToLower(side),
		OrderType: kraken.OrderTypeMarket,
		ClOrdID:   strings.ReplaceAll(uuid.NewString(), "-", ""),
	}

	var amount decimal.Decimal
	switch req.Type {
	case kraken.OrderSideSell:
		balance, err := o.GetCurrencyBalance(ctx, rule.BaseCurrency)
		if err != nil {
			if errors.Is(err, exchangeclient.ErrRateLimited) {
				return nil, exchangeclient.ErrSkipOrder
			}
			return nil, fmt.Errorf("get base currency balance %s: %w", rule.BaseCurrency, err)
		}
		bid, err := tick.BidPrice()
		if err != nil || !bid.IsPositive() {
			return nil, exchangeclient.ErrSkipOrder
		}

		req.Volume = balance.RoundDown(int32(rule.AmountPrecision)) //nolint:gosec
		if req.Volume.LessThan(minVolume) || req.Volume.Mul(bid).LessThan(minCost) {
			return nil, exchangeclient.ErrInsufficientBalance
		}
		amount = req.Volume
	case kraken.OrderSideBuy:
		balance, err := o.GetCurrencyBalance(ctx, rule.QuoteCurrency)
		if err != nil {
			if errors.Is(err, exchangeclient.ErrRateLimited) {
				return nil, exchangeclient.ErrSkipOrder
			}
			return nil, fmt.Errorf("get quote currency balance %s: %w", rule.QuoteCurrency, err)
		}
		ask, err := tick.AskPrice()
		if err != nil || !ask.IsPositive() {
			return nil, exchangeclient.ErrSkipOrder
		}

		spend := balance.Mul(decimal.NewFromInt(1).Sub(buySlippageBuffer))
		req.Volume = spend.Div(ask).RoundDown(int32(rule.AmountPrecision)) //nolint:gosec
		if req.Volume.LessThan(minVolume) || req.Volume.Mul(ask).LessThan(minCost) {
			return nil, exchangeclient.ErrInsufficientBalance
		}
		// Fee is taken from the bought currency, so the whole spend goes into the order.
		req.OFlags = []string{requests.OFlagFeeInBase}
		amount = req.Volume.Mul(ask).RoundDown(int32(rule.ValuePrecision)) //nolint:gosec
	default:
		return nil, fmt.Errorf("unsupported order side %s", side)
	}

	order, err := o.exClient.Spot().AddOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("place spot order: %w", err)
	}

	return &models.ExchangeOrderDTO{
		ExchangeOrderID: order.TxID[0],
		ClientOrderID:   req.ClOrdID,
		Amount:          amount,
	}, nil
}

func findTicker(tickers responses.TickerResponse, ticker string) (responses.Ticker, bool) {
	if t, ok := tickers[ticker]; ok {
		return t, true
	}
	if len(tickers) == 1 {
		for _, t := range tickers {
			return t, true
		}
	}
	return responses.Ticker{}, false
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
	if args.ExternalOrderID == nil || *args.ExternalOrderID == "" {
		return order, fmt.Errorf("external order id (kraken txid) is required")
	}
	txID := *args.ExternalOrderID

	res, err := o.exClient.Spot().QueryOrders(ctx, &requests.QueryOrdersRequest{TxIDs: []string{txID}})
	if err != nil {
		return nil, fmt.Errorf("query order: %w", err)
	}
	krakenOrder, ok := res[txID]
	if !ok {
		return nil, fmt.Errorf("order %s: %w", txID, kraken.ErrUnknownOrder)
	}

	switch krakenOrder.Status {
	case responses.OrderStatusClosed:
		order.State = models.ExchangeOrderStatusCompleted
	case responses.OrderStatusCanceled, responses.OrderStatusExpired:
		order.State = models.ExchangeOrderStatusFailed
		if krakenOrder.Reason != nil {
			order.FailReason = *krakenOrder.Reason
		}
	default:
		order.State = models.ExchangeOrderStatusInProgress
	}

	order.Amount = krakenOrder.VolExec

	info, err := o.pairInfo(ctx, *args.InstrumentID)
	if err != nil {
		return nil, fmt.Errorf("get pair %s: %w", *args.InstrumentID, err)
	}

	if info.quote == models.CurrencyCodeUSDT || info.quote == models.CurrencyCodeUSDC {
		order.AmountUSD = krakenOrder.Cost
		return order, nil
	}

	amountUSD, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
		Source: models.ExchangeSlugKraken.String(),
		From:   info.quote,
		To:     models.CurrencyCodeUSDT,
		Amount: krakenOrder.Cost.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("convert %s to usdt: %w", info.quote, err)
	}
	order.AmountUSD = amountUSD.Round(4)

	return order, nil
}
