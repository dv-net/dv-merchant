package clients_test

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/clients"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget/requests"
)

func Test_BitGetBaseClient(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Requires API keys, which are not available in CI")
	}
	bu, err := url.Parse("https://api.bitget.com")
	require.NoError(t, err)
	require.NotNil(t, bu)
	bc, err := clients.NewBaseClient(&clients.ClientOptions{
		AccessKey:  "",
		PassPhrase: "",
		SecretKey:  "",
		BaseURL:    bu,
	}, memory.NewStore())
	require.NoError(t, err)
	require.NotNil(t, bc)
	require.NotNil(t, bc.Signer())

	t.Run("Bitget Client", func(t *testing.T) {
		// 	t.Run("Account", func(t *testing.T) {
		t.Run("AccountAssets", func(t *testing.T) {
			res, err := bc.Spot().Account().AccountAssets(context.Background(), &requests.AccountAssetsRequest{
				Coin:      "BCH",
				AssetType: "all",
			})
			require.NoError(t, err)
			require.NotNil(t, res)
			for _, data := range res.Data {
				t.Logf("AccountAssets: %v", data)
			}
		})
		// t.Run("DepositAddress", func(t *testing.T) {
		// 	res, err := bc.Spot().Account().DepositAddress(context.Background(), &requests.DepositAddressRequest{
		// 		Coin:  "USDT",
		// 		Chain: "BEP20",
		// 	})
		// 	require.NoError(t, err)
		// 	require.NotNil(t, res)
		// 	t.Logf("DepositAddress: %v", res.Data)
		// })
		// t.Run("DepositRecords", func(t *testing.T) {
		// 	res, err := bc.Spot().Account().DepositRecords(context.Background(), &requests.DepositRecordsRequest{
		// 		Coin:      "TRX",
		// 		StartTime: strconv.FormatInt(time.Now().Add(-time.Hour*48).UnixMilli(), 10),
		// 		EndTime:   strconv.FormatInt(time.Now().UnixMilli(), 10),
		// 	})
		// 	require.NoError(t, err)
		// 	require.NotNil(t, res)
		// 	// for _, data := range res.Data {
		// 	// t.Logf("DepositRecords: %v", data)
		// 	// }
		// })
		// t.Run("WithdrawalRecords", func(t *testing.T) {
		// 	res, err := bc.Spot().Account().WithdrawalRecords(context.Background(), &requests.WithdrawalRecordsRequest{
		// 		Coin:      "USDT",
		// 		StartTime: strconv.FormatInt(time.Now().Add(-time.Hour*48).UnixMilli(), 10),
		// 		EndTime:   strconv.FormatInt(time.Now().UnixMilli(), 10),
		// 	})
		// 	require.NoError(t, err)
		// 	require.NotNil(t, res)
		// 	// for _, data := range res.Data {
		// 	// t.Logf("WithdrawalRecords: %v", data)
		// 	// }
		// })
		// t.Run("Withdrawal", func(t *testing.T) {
		// 	res, err := bc.Spot().Account().WalletWithdrawal(context.Background(), &requests.WalletWithdrawalRequest{
		// 		Coin:         "USDT",
		// 		Chain:        "BEP20",
		// 		TransferType: "on_chain",
		// 		Address:      "0x5A77B1806488107158248Bab379CAd2039B4aE70",
		// 		Size:         "10",
		// 	})
		// 	require.NoError(t, err)
		// 	require.NotNil(t, res)
		// 	t.Logf("Withdrawal: %v", res.Data)
		// })
	})
	// 	t.Run("Common", func(t *testing.T) {
	// 		t.Run("AllAccountBalance", func(t *testing.T) {
	// 			res, err := bc.Common().AllAccountBalance(context.Background())
	// 			require.NoError(t, err)
	// 			require.NotNil(t, res)
	// 			for _, data := range res.Data {
	// 				t.Logf("AllAccountBalance: %v", data)
	// 			}
	// 		})
	// 	})
	// t.Run("Trade", func(t *testing.T) {
	// 	t.Run("OrderInformation", func(t *testing.T) {
	// 		res, err := bc.Spot().Trade().OrderInformation(context.Background(), &requests.OrderInformationRequest{
	// 			OrderID: "1280730050351634",
	// 		})
	// 		require.NoError(t, err)
	// 		require.NotNil(t, res)
	// 		t.Logf("OrderInformation: %v", res.Data[0])
	// 	})
	// })
	// 	t.Run("Market", func(t *testing.T) {
	// 		t.Run("TickerInformation", func(t *testing.T) {
	// 			res, err := bc.Spot().Market().TickerInformation(context.Background(), &requests.TickerInformationRequest{})
	// 			require.NoError(t, err)
	// 			require.NotNil(t, res)
	// 			// for _, data := range res.Data {
	// 			// t.Logf("GetTickerInformation: %v", data)
	// 			// }
	// 		})
	// t.Run("CoinInformation", func(t *testing.T) {
	// 	res, err := bc.Spot().Market().CoinInformation(context.Background(), &requests.CoinInformationRequest{
	// 		Coin: "USDT",
	// 	})
	// 	require.NoError(t, err)
	// 	require.NotNil(t, res)
	// 	for _, data := range res.Data[0].Chains {
	// 		t.Logf("Test_BitGetBaseClient_Market_GetCoinInformation: %v", data)
	// 	}
	// })

	// 		t.Run("SymbolInformation", func(t *testing.T) {
	// 			res, err := bc.Spot().Market().SymbolInformation(context.Background(), &requests.SymbolInformationRequest{})
	// 			require.NoError(t, err)
	// 			require.NotNil(t, res)
	// 			// for _, data := range res.Data {
	// 			// 	t.Logf("SymbolInformation: %v", data)
	// 			// }
	// 		})
	// 	})
	// })
}
