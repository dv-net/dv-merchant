package kraken_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/responses"
)

func TestWallet_GetDepositMethods(t *testing.T) {
	client := newTestClient(t, "wallet-dm", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/0/private/DepositMethods", r.URL.Path)
		require.Equal(t, "USDT", readForm(t, r).Get("asset"))

		_, _ = w.Write([]byte(`{"error":[],"result":[
			{"method":"Tether USD (TRC20)","limit":false,"fee":"0.00000000","gen-address":true,"minimum":"1.00"},
			{"method":"Tether USD (ERC20)","limit":"100000.00","fee":"0","address-setup-fee":"5.00","gen-address":false,"minimum":"5.00"}
		]}`))
	})

	res, err := client.Wallet().GetDepositMethods(t.Context(), &requests.DepositMethodsRequest{Asset: "USDT"})
	require.NoError(t, err)
	require.Len(t, res, 2)

	trc, ok := res.Find("Tether USD (TRC20)")
	require.True(t, ok)
	require.False(t, trc.Limit.Limited)
	require.True(t, trc.GenAddress)
	require.True(t, decimal.NewFromInt(1).Equal(trc.Minimum))

	erc, ok := res.Find("Tether USD (ERC20)")
	require.True(t, ok)
	require.True(t, erc.Limit.Limited)
	require.True(t, decimal.NewFromInt(100000).Equal(erc.Limit.Value))
	require.True(t, decimal.NewFromInt(5).Equal(erc.AddressSetupFee))

	_, ok = res.Find("Tether USD (SOL)")
	require.False(t, ok)
}

func TestWallet_GetDepositAddresses(t *testing.T) {
	client := newTestClient(t, "wallet-da", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/0/private/DepositAddresses", r.URL.Path)
		form := readForm(t, r)
		require.Equal(t, "XRP", form.Get("asset"))
		require.Equal(t, "Ripple XRP", form.Get("method"))
		require.Equal(t, "true", form.Get("new"))

		_, _ = w.Write([]byte(`{"error":[],"result":[
			{"address":"rLHzPsX6oXkzU2qL12kHCH8G8cnZv1rBJh","expiretm":"0","new":true,"tag":"123456"},
			{"address":"rLHzPsX6oXkzU2qL12kHCH8G8cnZv1rBJh","expiretm":1700000000,"tag":"654321"}
		]}`))
	})

	res, err := client.Wallet().GetDepositAddresses(t.Context(), &requests.DepositAddressesRequest{
		Asset: "XRP", Method: "Ripple XRP", New: true,
	})
	require.NoError(t, err)
	require.Len(t, res, 2)
	require.True(t, res[0].New)
	require.Equal(t, "123456", res[0].Tag)
	require.Equal(t, responses.UnixTime(0), res[0].ExpireTm)
	require.Equal(t, responses.UnixTime(1700000000), res[1].ExpireTm)
}

func TestWallet_GetWithdrawMethods(t *testing.T) {
	client := newTestClient(t, "wallet-wm", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/0/private/WithdrawMethods", r.URL.Path)
		require.Equal(t, "USDT", readForm(t, r).Get("asset"))

		_, _ = w.Write([]byte(`{"error":[],"result":[
			{"asset":"USDT","method":"Tether USD (TRC20)","network":"Tron","minimum":"5.00"}
		]}`))
	})

	res, err := client.Wallet().GetWithdrawMethods(t.Context(), &requests.WithdrawMethodsRequest{Asset: "USDT"})
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "Tron", res[0].Network)
	require.True(t, decimal.NewFromInt(5).Equal(res[0].Minimum))
}

func TestWallet_GetWithdrawAddresses(t *testing.T) {
	verified := true
	client := newTestClient(t, "wallet-wa", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/0/private/WithdrawAddresses", r.URL.Path)
		form := readForm(t, r)
		require.Equal(t, "USDT", form.Get("asset"))
		require.Equal(t, "true", form.Get("verified"))

		_, _ = w.Write([]byte(`{"error":[],"result":[
			{"address":"TXyz1111111111111111111111111111aB","asset":"USDT","method":"Tether USD (TRC20)","key":"unverified-tron","verified":false},
			{"address":"TXyz1111111111111111111111111111aB","asset":"USDT","method":"Tether USD (TRC20)","key":"my-tron","verified":true},
			{"address":"0xAbCdEf0000000000000000000000000000000001","asset":"USDT","method":"Tether USD (ERC20)","key":"my-eth","verified":true}
		]}`))
	})

	res, err := client.Wallet().GetWithdrawAddresses(t.Context(), &requests.WithdrawAddressesRequest{
		Asset: "USDT", Verified: &verified,
	})
	require.NoError(t, err)
	require.Len(t, res, 3)

	t.Run("skips unverified", func(t *testing.T) {
		a, ok := res.FindByAddress("TXyz1111111111111111111111111111aB", "Tether USD (TRC20)")
		require.True(t, ok)
		require.Equal(t, "my-tron", a.Key)
	})
	t.Run("non-EVM address is case sensitive", func(t *testing.T) {
		_, ok := res.FindByAddress("txyz1111111111111111111111111111ab", "")
		require.False(t, ok)
	})
	t.Run("EVM address is case insensitive", func(t *testing.T) {
		a, ok := res.FindByAddress("0xabcdef0000000000000000000000000000000001", "")
		require.True(t, ok)
		require.Equal(t, "my-eth", a.Key)
	})
	t.Run("method must match", func(t *testing.T) {
		_, ok := res.FindByAddress("0xabcdef0000000000000000000000000000000001", "Tether USD (TRC20)")
		require.False(t, ok)
	})
}

func TestWallet_GetWithdrawInfo(t *testing.T) {
	client := newTestClient(t, "wallet-wi", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/0/private/WithdrawInfo", r.URL.Path)
		form := readForm(t, r)
		require.Equal(t, "USDT", form.Get("asset"))
		require.Equal(t, "my-tron", form.Get("key"))
		require.Equal(t, "100.5", form.Get("amount"))

		_, _ = w.Write([]byte(`{"error":[],"result":{"method":"Tether USD (TRC20)","limit":"1500.00","amount":"97.50","fee":"3.00"}}`))
	})

	res, err := client.Wallet().GetWithdrawInfo(t.Context(), &requests.WithdrawInfoRequest{
		Asset: "USDT", Key: "my-tron", Amount: decimal.RequireFromString("100.5"),
	})
	require.NoError(t, err)
	require.Equal(t, "Tether USD (TRC20)", res.Method)
	require.True(t, decimal.RequireFromString("1500").Equal(res.Limit))
	require.True(t, decimal.RequireFromString("97.5").Equal(res.Amount))
	require.True(t, decimal.RequireFromString("3").Equal(res.Fee))
}

func TestWallet_Withdraw(t *testing.T) {
	client := newTestClient(t, "wallet-w", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/0/private/Withdraw", r.URL.Path)
		form := readForm(t, r)
		require.Equal(t, "USDT", form.Get("asset"))
		require.Equal(t, "my-tron", form.Get("key"))
		require.Equal(t, "100", form.Get("amount"))
		require.Equal(t, "TXyz1111111111111111111111111111aB", form.Get("address"))
		require.Equal(t, "3.5", form.Get("max_fee"))

		_, _ = w.Write([]byte(`{"error":[],"result":{"refid":"FTQcuak-V6Za8qrWnhzTx67yYHz8Tg"}}`))
	})

	res, err := client.Wallet().Withdraw(t.Context(), &requests.WithdrawRequest{
		Asset:   "USDT",
		Key:     "my-tron",
		Amount:  decimal.NewFromInt(100),
		Address: "TXyz1111111111111111111111111111aB",
		MaxFee:  decimal.RequireFromString("3.5"),
	})
	require.NoError(t, err)
	require.Equal(t, "FTQcuak-V6Za8qrWnhzTx67yYHz8Tg", res.RefID)
}

func TestWallet_Withdraw_OptionalParams(t *testing.T) {
	client := newTestClient(t, "wallet-w-opt", func(w http.ResponseWriter, r *http.Request) {
		form := readForm(t, r)
		require.False(t, form.Has("address"))
		require.False(t, form.Has("max_fee"))
		_, _ = w.Write([]byte(`{"error":[],"result":{"refid":"REF"}}`))
	})

	_, err := client.Wallet().Withdraw(t.Context(), &requests.WithdrawRequest{
		Asset: "USDT", Key: "my-tron", Amount: decimal.NewFromInt(1),
	})
	require.NoError(t, err)
}

func TestWallet_Withdraw_MissingRefID(t *testing.T) {
	client := newTestClient(t, "wallet-w-noref", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"error":[],"result":{}}`))
	})

	_, err := client.Wallet().Withdraw(t.Context(), &requests.WithdrawRequest{
		Asset: "USDT", Key: "my-tron", Amount: decimal.NewFromInt(1),
	})
	require.ErrorContains(t, err, "without refid")
}

func TestWallet_InvalidRequests(t *testing.T) {
	client := newTestClient(t, "wallet-invalid", func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("request must not be sent")
	})
	ctx := t.Context()
	w := client.Wallet()

	checks := map[string]func() error{
		"deposit methods nil": func() error { _, err := w.GetDepositMethods(ctx, nil); return err },
		"deposit methods no asset": func() error {
			_, err := w.GetDepositMethods(ctx, &requests.DepositMethodsRequest{})
			return err
		},
		"deposit addresses no method": func() error {
			_, err := w.GetDepositAddresses(ctx, &requests.DepositAddressesRequest{Asset: "USDT"})
			return err
		},
		"withdraw info no key": func() error {
			_, err := w.GetWithdrawInfo(ctx, &requests.WithdrawInfoRequest{Asset: "USDT", Amount: decimal.NewFromInt(1)})
			return err
		},
		"withdraw info zero amount": func() error {
			_, err := w.GetWithdrawInfo(ctx, &requests.WithdrawInfoRequest{Asset: "USDT", Key: "k"})
			return err
		},
		"withdraw nil": func() error { _, err := w.Withdraw(ctx, nil); return err },
		"withdraw no key": func() error {
			_, err := w.Withdraw(ctx, &requests.WithdrawRequest{Asset: "USDT", Amount: decimal.NewFromInt(1)})
			return err
		},
		"withdraw negative amount": func() error {
			_, err := w.Withdraw(ctx, &requests.WithdrawRequest{Asset: "USDT", Key: "k", Amount: decimal.NewFromInt(-1)})
			return err
		},
		"withdraw negative max fee": func() error {
			_, err := w.Withdraw(ctx, &requests.WithdrawRequest{
				Asset: "USDT", Key: "k", Amount: decimal.NewFromInt(1), MaxFee: decimal.NewFromInt(-1),
			})
			return err
		},
	}
	for name, check := range checks {
		t.Run(name, func(t *testing.T) {
			require.ErrorIs(t, check(), kraken.ErrInvalidWalletRequest)
		})
	}
}

func TestWallet_GetWithdrawStatus(t *testing.T) {
	start := time.Unix(1688000000, 0)
	client := newTestClient(t, "wallet-ws", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/0/private/WithdrawStatus", r.URL.Path)
		form := readForm(t, r)
		require.Equal(t, "USDT", form.Get("asset"))
		require.Equal(t, "1688000000", form.Get("start"))
		require.False(t, form.Has("end"))

		_, _ = w.Write([]byte(`{"error":[],"result":[
			{"method":"Tether USD (TRC20)","network":"Tron","aclass":"currency","asset":"USDT",
			 "refid":"REF-OK","txid":"abc123","info":"TXyz","amount":"97.00","fee":"3.00",
			 "time":1688665496,"status":"Success","key":"my-tron"},
			{"method":"Tether USD (TRC20)","aclass":"currency","asset":"USDT",
			 "refid":"REF-PENDING","txid":"","info":"TXyz","amount":"10","fee":"3",
			 "time":1688665500,"status":"Pending"},
			{"method":"Tether USD (TRC20)","aclass":"currency","asset":"USDT",
			 "refid":"REF-SETTLED","txid":"def456","info":"TXyz","amount":"10","fee":"3",
			 "time":1688665501,"status":"Settled"},
			{"method":"Tether USD (TRC20)","aclass":"currency","asset":"USDT",
			 "refid":"REF-CANCELED","txid":"","info":"TXyz","amount":"10","fee":"3",
			 "time":1688665502,"status":"Pending","status-prop":"canceled"},
			{"method":"Tether USD (TRC20)","aclass":"currency","asset":"USDT",
			 "refid":"REF-FAIL","txid":"","info":"TXyz","amount":"10","fee":"3",
			 "time":1688665503,"status":"Failure"},
			{"method":"Tether USD (TRC20)","aclass":"currency","asset":"USDT",
			 "refid":"REF-HOLD","txid":"","info":"TXyz","amount":"10","fee":"3",
			 "time":1688665504,"status":"Pending","status-prop":"onhold"}
		]}`))
	})

	res, err := client.Wallet().GetWithdrawStatus(t.Context(), &requests.WithdrawStatusRequest{Asset: "USDT", Start: start})
	require.NoError(t, err)
	require.Len(t, res, 6)

	ok, found := res.FindByRefID("REF-OK")
	require.True(t, found)
	require.True(t, ok.IsSuccess())
	require.Equal(t, "abc123", ok.TxID)
	require.Equal(t, responses.UnixTime(1688665496), ok.Time)
	require.True(t, decimal.NewFromInt(97).Equal(ok.Amount))

	cases := map[string]struct{ success, failed, inProgress bool }{
		"REF-PENDING":  {inProgress: true},
		"REF-SETTLED":  {inProgress: true},
		"REF-CANCELED": {failed: true},
		"REF-FAIL":     {failed: true},
		"REF-HOLD":     {inProgress: true},
	}
	for refID, want := range cases {
		wd, found := res.FindByRefID(refID)
		require.True(t, found, refID)
		require.Equal(t, want.success, wd.IsSuccess(), refID)
		require.Equal(t, want.failed, wd.IsFailed(), refID)
		require.Equal(t, want.inProgress, wd.IsInProgress(), refID)
	}

	_, found = res.FindByRefID("NOPE")
	require.False(t, found)
}

func TestWallet_ErrorMapping(t *testing.T) {
	cases := map[string]error{
		"EFunding:Unknown withdraw key": exchangeclient.ErrWithdrawalAddressNotWhitelisted,
		"EFunding:Insufficient funds":   exchangeclient.ErrInsufficientBalance,
		"EFunding:Invalid amount":       exchangeclient.ErrMinWithdrawalBalance,
		"EFunding:Max fee exceeded":     kraken.ErrMaxFeeExceeded,
		"EFunding:Too many addresses":   kraken.ErrTooManyAddresses,
		"EFunding:Unknown reference id": kraken.ErrUnknownReference,
	}
	for krakenErr, want := range cases {
		t.Run(krakenErr, func(t *testing.T) {
			client := newTestClient(t, "wallet-err-"+krakenErr, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"error":["` + krakenErr + `"]}`))
			})
			_, err := client.Wallet().Withdraw(t.Context(), &requests.WithdrawRequest{
				Asset: "USDT", Key: "k", Amount: decimal.NewFromInt(1),
			})
			require.ErrorIs(t, err, want)
		})
	}
}
