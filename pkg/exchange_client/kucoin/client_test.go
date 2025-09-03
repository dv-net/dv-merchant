package kucoin

import (
	"net/url"
	"os"
	"testing"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/requests"
	"github.com/dv-net/dv-merchant/pkg/key_value"

	"github.com/stretchr/testify/require"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func Test_Account(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Requires API keys, which are not available in CI")
	}
	bURL, err := url.Parse("https://api.kucoin.com")
	require.NoError(t, err)
	client := NewBaseClient(&ClientOptions{
		KeyAPI:        "",
		KeySecret:     "",
		KeyPassphrase: "",
		BaseURL:       bURL,
	}, memory.NewStore(), key_value.NewInMemory())
	require.NotNil(t, client)
	t.Run("GetAPIKeyInfo", func(t *testing.T) {
		keyInfo, err := client.Account().GetAPIKeyInfo(t.Context(), requests.GetAPIKeyInfo{})
		require.NoError(t, err)
		require.NotNil(t, keyInfo)
		t.Logf("keyInfo: %+v", keyInfo.Info)
	})
	t.Run("GetAccountList", func(t *testing.T) {
		accountList, err := client.Account().GetAccountList(t.Context(), requests.GetAccountList{})
		require.NoError(t, err)
		require.NotNil(t, accountList)
		if len(accountList.Accounts) != 0 {
			t.Logf("accountList: %+v", accountList.Accounts[0])
		}
	})
	t.Run("GetCurrencyList", func(t *testing.T) {
		currencyList, err := client.Market().GetCurrencyList(t.Context(), requests.GetCurrencyList{})
		require.NoError(t, err)
		require.NotNil(t, currencyList)
		t.Logf("currencyList: %+v", currencyList.Currencies[0])
		t.Logf("currencyList: %+v", currencyList.Currencies[0].Chains[0])
	})
	t.Run("GetAllSymbols", func(t *testing.T) {
		allSymbols, err := client.Public().GetAllSymbols(t.Context(), requests.GetAllSymbols{})
		require.NoError(t, err)
		require.NotNil(t, allSymbols)
		t.Logf("allSymbols: %d", len(allSymbols.Symbols))
		t.Logf("allSymbols: %+v", allSymbols.Symbols[0])
	})
	t.Run("GetDepositAddresses", func(t *testing.T) {
		depositAddress, err := client.Account().GetDepositAddress(t.Context(), requests.GetDepositAddress{
			Currency: "USDT",
			Chain:    "trc20",
		})
		require.NoError(t, err)
		require.NotNil(t, depositAddress)
		t.Logf("depositAddress: %+v", depositAddress.Addresses)
	})
}
