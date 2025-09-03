package gateio_test

import (
	"net/url"
	"os"
	"testing"

	gateio "github.com/dv-net/dv-merchant/pkg/exchange_client/gate"
	"github.com/stretchr/testify/require"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func Test_BaseClient(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping Gate.io tests in CI environment")
	}

	bURL, err := url.Parse("https://api.gateio.ws")
	require.NoError(t, err)

	client, err := gateio.NewBaseClient(&gateio.ClientOptions{
		AccessKey: "",
		SecretKey: "",
		BaseURL:   bURL,
	}, memory.NewStore())
	require.NoError(t, err)

	// t.Run("AccountClient", func(t *testing.T) {
	// 	accountClient := client.Account()
	// 	require.NotNil(t, accountClient)

	// 	t.Run("GetAccountDetail", func(t *testing.T) {
	// 		detail, err := accountClient.GetAccountDetail(t.Context())
	// 		require.NoError(t, err)
	// 		require.NotNil(t, detail)
	// 	})
	// })

	// t.Run("WalletClient", func(t *testing.T) {
	// 	walletClient := client.Wallet()
	// 	require.NotNil(t, walletClient)

	// 	t.Run("GetCurrencyChain", func(t *testing.T) {
	// 		chains, err := walletClient.GetCurrencyChainsSupported(t.Context(), &gateio.GetCurrencyChainsSupportedRequest{
	// 			Currency: "BTC",
	// 		})
	// 		require.NoError(t, err)
	// 		require.NotNil(t, chains)
	// 	})
	// })

	t.Run("SpotClient", func(t *testing.T) {
		spotClient := client.Spot()
		require.NotNil(t, spotClient)

		// 	t.Run("GetSpotAccountBalances", func(t *testing.T) {
		// 		balances, err := spotClient.GetSpotAccountBalances(t.Context(), &gateio.GetSpotAccountBalancesRequest{
		// 			Currency: "BTC",
		// 		})
		// 		require.NoError(t, err)
		// 		require.NotNil(t, balances)
		// 		require.Greater(t, len(balances.Data), 0, "Expected at least one balance")
		// 		for _, balance := range balances.Data {
		// 			t.Logf("Currency: %s, Available: %s, Locked: %s", balance.Currency, balance.Available, balance.Locked)
		// 		}
		// 	})
		// t.Run("GetSpotCurrencies", func(t *testing.T) {
		// 	currencies, err := spotClient.GetSpotCurrencies(t.Context())
		// 	require.NoError(t, err)
		// 	require.NotNil(t, currencies)
		// })

		// t.Run("GetSpotCurrency", func(t *testing.T) {
		// 	currency, err := spotClient.GetSpotCurrency(t.Context(), "BTC")
		// 	require.NoError(t, err)
		// 	require.NotNil(t, currency)
		// 	t.Logf("Currency: %+v\n", currency.Data)
		// })

		// t.Run("GetSpotSupportedCurrencyPairs", func(t *testing.T) {
		// 	pairs, err := spotClient.GetSpotSupportedCurrencyPairs(t.Context())
		// 	require.NoError(t, err)
		// 	require.NotNil(t, pairs)
		// 	t.Logf("First pair: %+v\n", pairs.Data[0])
		// })

		// 	t.Run("GetSpotSupportedCurrencyPair", func(t *testing.T) {
		// 		pair, err := spotClient.GetSpotSupportedCurrencyPair(t.Context(), "BTC_USDT")
		// 		require.NoError(t, err)
		// 		require.NotNil(t, pair)
		// 	})

		// 	t.Run("GetTickersInfo", func(t *testing.T) {
		// 		t.Run("Single ticker", func(t *testing.T) {
		// 			tickers, err := spotClient.GetTickersInfo(t.Context(), &gateio.GetTickersInfoRequest{
		// 				CurrencyPair: "BTC_USDT",
		// 			})
		// 			require.NoError(t, err)
		// 			require.NotNil(t, tickers.Data)
		// 			require.Greater(t, len(tickers.Data), 0, "Expected at least one ticker for BTC_USDT")
		// 		})
		// 		t.Run("All tickers", func(t *testing.T) {
		// 			tickers, err := spotClient.GetTickersInfo(t.Context(), &gateio.GetTickersInfoRequest{})
		// 			require.NoError(t, err)
		// 			require.NotNil(t, tickers)
		// 			require.Greater(t, len(tickers.Data), 0, "Expected at least one ticker")
		// 		})
		// 	})
	})

	t.Run("WalletClient", func(t *testing.T) {
		walletClient := client.Wallet()
		require.NotNil(t, walletClient)
		// t.Run("GetDepositAddress", func(t *testing.T) {
		// 	address, err := walletClient.GetDepositAddress(t.Context(), &gateio.GetDepositAddressRequest{
		// 		Currency: "TON",
		// 	})
		// 	require.NoError(t, err)
		// 	require.NotNil(t, address)
		// 	require.NotEmpty(t, address.Data.Address, "Expected a valid deposit address")
		// 	t.Logf("Deposit Address: %+v", address.Data)
		// })

		// Withdrawal History: &{ID:w79417392 Currency:USDT Address:0x5a77b1806488107158248bab379cad2039b4ae70 Amount:132.42180703 Fee:0.5 Txid:0xf5d4fa4f425f8dda802dd55a391df6ac28a9a18b699321264d9ac62f683514cf Chain:BSC Status:DONE WithdrawOrderID: BlockNumber:52029189 FailReason:}
		// t.Run("GetWithdrawalHistory", func(t *testing.T) {
		// 	history, err := walletClient.GetWithdrawalHistory(t.Context(), &gateio.GetWithdrawalHistoryRequest{
		// 		Limit:        "1",
		// 		WithdrawalID: "w79417392",
		// 	})
		// 	require.NoError(t, err)
		// 	require.NotNil(t, history)
		// 	t.Logf("Withdrawal History: %+v", history.Data[0])
		// })
		// t.Run("GetWithdrawalRules", func(t *testing.T) {
		// 	rules, err := walletClient.GetWithdrawalRules(t.Context(), &gateio.GetWithdrawalRulesRequest{
		// 		Currency: "USDT",
		// 	})
		// 	require.NoError(t, err)
		// 	require.NotNil(t, rules)
		// 	for _, rule := range rules.Data {
		// 		t.Logf("Withdrawal Rule: %+v\n", rule)
		// 	}
		// })
	})
}
