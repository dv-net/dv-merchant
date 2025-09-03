package htx

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	htxrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/htx/requests"
)

const (
	MainBaseURL = "https://api.huobi.pro"
)

func TestNewHtxClient(t *testing.T) {
	t.Parallel()
	if os.Getenv("CI") == "true" {
		t.Skip("Requires API keys, which are not available in CI")
	}
	ctx := context.Background()
	baseURL, err := url.Parse(MainBaseURL)
	require.NoError(t, err)
	c, err := NewBaseClient(&ClientOptions{
		AccessKey: "",
		SecretKey: "",
		BaseURL:   baseURL,
	})
	require.NoError(t, err)
	require.NotNil(t, c)
	t.Run("Account", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, c.Account())
		t.Run("GetAllAccounts", func(t *testing.T) {
			accs, err := c.Account().GetAllAccounts(ctx)
			require.NoError(t, err)
			require.NotNil(t, accs)
		})
		t.Run("GetAccountBalance", func(t *testing.T) {
			accs, err := c.Account().GetAllAccounts(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, accs)

			for _, acc := range accs {
				t.Run(fmt.Sprintf("GetAccountBalance:%d", acc.ID), func(t *testing.T) {
					bal, err := c.Account().GetAccountBalance(ctx, acc)
					require.NoError(t, err)
					require.NotNil(t, bal)
					t.Log(strings.ToUpper(bal.Type.String())+" currency count:", len(bal.List))
				})
			}
		})
	})
	t.Run("Market", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, c.Market())
		t.Run("GetMarketTickers", func(t *testing.T) {
			tickers, err := c.Market().GetMarketTickers(ctx)
			require.NoError(t, err)
			require.NotNil(t, tickers)
			t.Log("symbols count:", len(tickers))
		})
		t.Run("GetMarketDetails", func(t *testing.T) {
			ticker, err := c.Market().GetMarketDetails(ctx, "btcusdt")
			require.NoError(t, err)
			require.NotNil(t, ticker)
		})
	})
	t.Run("Order", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, c.Order())
		t.Run("GetOrderHistory", func(t *testing.T) {
			orders, err := c.Order().GetOrdersHistory(ctx, nil)
			require.NoError(t, err)
			require.NotNil(t, orders)
			t.Log("orders count:", len(orders))
		})
	})
	t.Run("User", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, c.User())
		t.Run("GetAPIKeyInformation", func(t *testing.T) {
			accounts, err := c.Account().GetAllAccounts(ctx)
			require.NoError(t, err)
			require.NotEmpty(t, accounts)
			accountID := accounts[0].ID
			_, err = c.User().GetAPIKeyInformation(ctx, &htxrequests.GetAPIKeyInformationRequest{
				AccessKey: c.AccessKey(),
				UID:       strconv.Itoa(int(accountID)),
			})
			require.Error(t, err)
		})
	})
	t.Run("Wallet", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, c.Wallet())
		t.Run("GetWithdrawalAddress", func(t *testing.T) {
			t.Run(fmt.Sprintf("GetWithdrawalAddress:%s", "usdt"), func(t *testing.T) {
				address, err := c.Wallet().GetWithdrawalAddress(ctx, &htxrequests.WithdrawalAddressRequest{
					Currency: "usdt",
				})
				require.NoError(t, err)
				t.Log(address)
			})
		})
		t.Run("GetDepositAddress", func(t *testing.T) {
			t.Run(fmt.Sprintf("GetDepositAddress:%s", "usdt"), func(t *testing.T) {
				address, err := c.Wallet().GetDepositAddress(ctx, &htxrequests.DepositAddressRequest{
					Currency: "usdt",
				})
				require.NoError(t, err)
				t.Log(address)
			})
		})
	})
}
