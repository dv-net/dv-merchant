package binance_test

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/binance"
	binancemodels "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/models"
	binancerequests "github.com/dv-net/dv-merchant/pkg/exchange_client/binance/requests"
)

func TestFastestEndpoint(t *testing.T) {
	t.Parallel()
	endpoints := []*url.URL{
		{Host: "api.binance.com", Scheme: "https"},
		{Host: "api-gcp.binance.com", Scheme: "https"},
		{Host: "api1.binance.com", Scheme: "https"},
		{Host: "api2.binance.com", Scheme: "https"},
		{Host: "api3.binance.com", Scheme: "https"},
		{Host: "api4.binance.com", Scheme: "https"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.Host, func(t *testing.T) {
			t.Parallel()
			bCl, err := binance.NewBaseClient(&binance.ClientOptions{
				BaseURL:      endpoint,
				PublicClient: true,
			})
			require.NoError(t, err)
			require.NotNil(t, bCl)
			now := time.Now()
			success, err := bCl.MarketData().GetPing(context.Background())
			require.NoError(t, err)
			require.True(t, success)
			diff := time.Since(now)
			t.Logf("Time taken for %s: %v", endpoint.Host, diff)
		})
	}
}

func TestURL(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Requires API keys, which are not available in CI")
	}
	t.Parallel()
	bURL, err := url.Parse("https://api.binance.com")
	require.NoError(t, err)

	bCl, err := binance.NewBaseClient(&binance.ClientOptions{
		APIKey:       "",
		SecretKey:    "",
		BaseURL:      bURL,
		PublicClient: false,
	})
	require.NoError(t, err)
	require.NotNil(t, bCl)

	require.NotNil(t, bCl.MarketData())
	require.NotNil(t, bCl.Wallet())
	require.NotNil(t, bCl.Spot())

	t.Run("GetDefaultDepositAddress", func(t *testing.T) {
		t.Parallel()
		res, err := bCl.Wallet().GetDefaultDepositAddress(context.Background(), &binancerequests.GetDefaultDepositAddressRequest{
			Coin:    "USDT",
			Network: "TRX",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	t.Run("GetDepositAddresses", func(t *testing.T) {
		t.Parallel()
		depositAddresses, err := bCl.Wallet().GetDepositAddresses(context.Background(), &binancerequests.GetDepositAddressesRequest{
			Coin: "BTC",
		})
		require.NoError(t, err)
		require.NotNil(t, depositAddresses)
	})

	t.Run("GetServerTime", func(t *testing.T) {
		t.Parallel()
		serverTime, err := bCl.MarketData().GetServerTime(context.Background())
		require.NoError(t, err)
		require.NotNil(t, serverTime)
	})

	t.Run("GetExchangeInfo", func(t *testing.T) {
		t.Parallel()
		exchangeInfo, err := bCl.MarketData().GetExchangeInfo(context.Background(), &binancerequests.GetExchangeInfoRequest{})
		require.NoError(t, err)
		require.NotNil(t, exchangeInfo)
	})

	t.Run("GetSymbolPriceTicker", func(t *testing.T) {
		t.Parallel()
		tickerPrice, err := bCl.MarketData().GetSymbolPriceTicker(context.Background(), &binancerequests.GetSymbolPriceTickerRequest{
			Symbol: "BTCUSDT",
		})
		require.NoError(t, err)
		require.NotNil(t, tickerPrice)
	})

	t.Run("GetSymbolsPriceTicker", func(t *testing.T) {
		t.Parallel()
		tickerPrices, err := bCl.MarketData().GetSymbolsPriceTicker(context.Background(), &binancerequests.GetSymbolsPriceTickerRequest{})
		require.NoError(t, err)
		require.NotNil(t, tickerPrices)
	})

	t.Run("GetAllCoinInformation", func(t *testing.T) {
		t.Parallel()
		coinInfo, err := bCl.Wallet().GetAllCoinInformation(context.Background())
		require.NoError(t, err)
		require.NotNil(t, coinInfo)
	})

	t.Run("GetFundingAssets", func(t *testing.T) {
		t.Parallel()
		fundingAssets, err := bCl.Wallet().GetFundingAssets(context.Background(), &binancerequests.GetFundingAssetsRequest{})
		require.NoError(t, err)
		require.NotNil(t, fundingAssets)
	})

	t.Run("GetUserBalances", func(t *testing.T) {
		t.Parallel()
		userBalances, err := bCl.Wallet().GetUserBalances(context.Background(), &binancerequests.GetUserBalancesRequest{})
		require.NoError(t, err)
		require.NotNil(t, userBalances)
	})

	t.Run("TestSpotOrder", func(t *testing.T) {
		t.Parallel()
		order, err := bCl.Spot().TestNewOrder(context.Background(), &binancerequests.TestNewOrderRequest{
			NewOrderRequest: binancerequests.NewOrderRequest{
				Symbol:        "BTCUSDT",
				Side:          "BUY",
				Type:          "MARKET",
				QuoteOrderQty: "10",
			},
			ComputeCommissionRates: true,
		})
		require.NoError(t, err)
		require.NotNil(t, order)
	})

	t.Run("TestMarketOrder", func(t *testing.T) {
		t.Parallel()
		order := &binancerequests.NewOrderRequest{
			Symbol:        "BTCUSDT",
			Side:          "BUY",
			Type:          "MARKET",
			QuoteOrderQty: "10",
		}
		require.NoError(t, err)
		response, err := bCl.Spot().NewOrder(context.Background(), order)
		require.NoError(t, err)
		require.NotNil(t, response)
	})

	t.Run("TestFakeMarketOrder", func(t *testing.T) {
		t.Parallel()
		order := &binancerequests.TestNewOrderRequest{
			NewOrderRequest: binancerequests.NewOrderRequest{
				Symbol:   "TRXUSDT",
				Side:     "BUY",
				Type:     "MARKET",
				Quantity: "0.1",
			},
		}
		require.NoError(t, err)
		response, err := bCl.Spot().TestNewOrder(context.Background(), order)
		require.NoError(t, err)
		require.NotNil(t, response)
	})

	t.Run("GetWithdrawalHistory", func(t *testing.T) {
		t.Parallel()
		withdrawalHistory, err := bCl.Wallet().GetWithdrawalHistory(context.Background(), &binancerequests.GetWithdrawalHistoryRequest{})
		require.NoError(t, err)
		require.NotNil(t, withdrawalHistory)
	})

	t.Run("UniversalTransfer", func(t *testing.T) {
		t.Parallel()
		transfer, err := bCl.Wallet().UniversalTransfer(context.Background(), &binancerequests.UniversalTransferRequest{
			Type:   binancemodels.TransferTypeSpotToFunding,
			Asset:  "TRX",
			Amount: decimal.NewFromFloat(9.999999999999).String(),
		})
		require.NoError(t, err)
		require.NotNil(t, transfer)
	})
}
