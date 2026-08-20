package mexc_test

import (
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/mexc"
)

const (
	testAPIKey    = ""
	testSecretKey = ""
)

func Test_BaseClient(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping MEXC tests in CI environment")
	}
	if testAPIKey == "" || testSecretKey == "" {
		t.Skip("testAPIKey/testSecretKey are not set")
	}

	bURL, err := url.Parse("https://api.mexc.com")
	require.NoError(t, err)

	client, err := mexc.NewBaseClient(&mexc.ClientOptions{
		APIKey:    testAPIKey,
		SecretKey: testSecretKey,
		BaseURL:   bURL,
	}, memory.NewStore())
	require.NoError(t, err)
	require.NotNil(t, client)

	t.Run("AccountClient", func(t *testing.T) {
		accountClient := client.Account()
		require.NotNil(t, accountClient)

		t.Run("GetAccountInfo", func(t *testing.T) {
			info, err := accountClient.GetAccountInfo(t.Context())
			require.NoError(t, err)
			require.NotNil(t, info)

			t.Logf("canTrade=%v canWithdraw=%v canDeposit=%v accountType=%s permissions=%v",
				info.CanTrade, info.CanWithdraw, info.CanDeposit, info.AccountType, info.Permissions)
			for _, balance := range info.Balances {
				if balance.Free.IsZero() && balance.Locked.IsZero() {
					continue
				}
				t.Logf("asset=%s free=%s locked=%s", balance.Asset, balance.Free.String(), balance.Locked.String())
			}
		})
	})
}
