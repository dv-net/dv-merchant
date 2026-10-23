package kraken_test

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken"
)

var testSecret = base64.StdEncoding.EncodeToString([]byte("test-secret"))

func newTestClient(t *testing.T, apiKey string, handler http.HandlerFunc) *kraken.BaseClient {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	bURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client, err := kraken.NewBaseClient(&kraken.ClientOptions{
		APIKey:    apiKey,
		SecretKey: testSecret,
		BaseURL:   bURL,
	}, memory.NewStore())
	require.NoError(t, err)
	return client
}

func TestAccount_GetBalance(t *testing.T) {
	client := newTestClient(t, "balance-key", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/0/private/Balance", r.URL.Path)
		require.Equal(t, "balance-key", r.Header.Get("API-Key"))
		require.NotEmpty(t, r.Header.Get("API-Sign"))
		require.Contains(t, r.Header.Get("Content-Type"), "application/x-www-form-urlencoded")

		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		form, err := url.ParseQuery(string(body))
		require.NoError(t, err)
		require.NotEmpty(t, form.Get("nonce"))

		_, _ = w.Write([]byte(`{"error":[],"result":{"ZUSD":"171288.6158","XXBT":"0.0000011","USDT.F":"12.5"}}`))
	})

	res, err := client.Account().GetBalance(t.Context())
	require.NoError(t, err)
	require.Len(t, res, 3)
	require.True(t, decimal.RequireFromString("171288.6158").Equal(res["ZUSD"]))
	require.True(t, decimal.RequireFromString("0.0000011").Equal(res["XXBT"]))
	require.True(t, decimal.RequireFromString("12.5").Equal(res["USDT.F"]))
}

func TestAccount_GetBalanceEx(t *testing.T) {
	client := newTestClient(t, "balance-ex-key", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/0/private/BalanceEx", r.URL.Path)
		_, _ = w.Write([]byte(`{"error":[],"result":{
			"ZUSD":{"balance":"25435.21","hold_trade":"8249.76"},
			"XXBT":{"balance":"1.2","hold_trade":"0","credit":"0.5","credit_used":"0.1"}
		}}`))
	})

	res, err := client.Account().GetBalanceEx(t.Context())
	require.NoError(t, err)
	require.Len(t, res, 2)
	require.True(t, decimal.RequireFromString("17185.45").Equal(res["ZUSD"].Available()))
	require.True(t, decimal.RequireFromString("1.6").Equal(res["XXBT"].Available()))
}

func TestAccount_ErrorMapping(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"invalid key", http.StatusOK, `{"error":["EAPI:Invalid key"]}`, exchangeclient.ErrInvalidAPICredentials},
		{"permission denied", http.StatusOK, `{"error":["EGeneral:Permission denied"]}`, exchangeclient.ErrIncorrectAPIPermissions},
		{"rate limit", http.StatusOK, `{"error":["EAPI:Rate limit exceeded"]}`, exchangeclient.ErrRateLimited},
		{"invalid nonce", http.StatusOK, `{"error":["EAPI:Invalid nonce"]}`, kraken.ErrInvalidNonce},
		{"html 429", http.StatusTooManyRequests, `<html>slow down</html>`, exchangeclient.ErrRateLimited},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(t, "err-"+tc.name, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})

			_, err := client.Account().GetBalance(t.Context())
			require.ErrorIs(t, err, tc.want)
		})
	}

	t.Run("unknown error", func(t *testing.T) {
		client := newTestClient(t, "err-unknown", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"error":["EGeneral:Internal error"]}`))
		})
		_, err := client.Account().GetBalance(t.Context())
		require.ErrorContains(t, err, "EGeneral:Internal error")
	})
}

func TestAccount_NoncesIncreaseUnderConcurrency(t *testing.T) {
	var (
		mu     sync.Mutex
		nonces []string
	)
	client := newTestClient(t, "concurrent-key", func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		mu.Lock()
		nonces = append(nonces, r.PostForm.Get("nonce"))
		mu.Unlock()
		_, _ = w.Write([]byte(`{"error":[],"result":{}}`))
	})

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			_, err := client.Account().GetBalance(t.Context())
			require.NoError(t, err)
		})
	}
	wg.Wait()

	require.Len(t, nonces, 8)
	for i := 1; i < len(nonces); i++ {
		prev, _ := decimal.NewFromString(nonces[i-1])
		cur, _ := decimal.NewFromString(nonces[i])
		require.True(t, cur.GreaterThan(prev), "nonces must arrive in increasing order: %v", nonces)
	}
}

const (
	liveAPIKey    = ""
	liveSecretKey = ""
)

func TestAccount_Live(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping Kraken tests in CI environment")
	}
	if liveAPIKey == "" || liveSecretKey == "" {
		t.Skip("liveAPIKey/liveSecretKey are not set")
	}

	bURL, err := url.Parse("https://api.kraken.com")
	require.NoError(t, err)

	client, err := kraken.NewBaseClient(&kraken.ClientOptions{
		APIKey:    liveAPIKey,
		SecretKey: liveSecretKey,
		BaseURL:   bURL,
	}, memory.NewStore())
	require.NoError(t, err)

	res, err := client.Account().GetBalanceEx(t.Context())
	require.NoError(t, err)
	for asset, b := range res {
		t.Logf("asset=%s balance=%s hold=%s available=%s", asset, b.Balance, b.HoldTrade, b.Available())
	}
}
