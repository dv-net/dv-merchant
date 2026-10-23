package exrate_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/dv-net/dv-merchant/internal/service/exrate"
	"github.com/dv-net/dv-merchant/pkg/logger"
	mxlogger "github.com/dv-net/mx/logger"
)

const krakenAssetPairsResp = `{"error":[],"result":{
	"XBTUSDT":{"altname":"XBTUSDT","wsname":"XBT/USDT","base":"XXBT","quote":"USDT","status":"online"},
	"XETHZUSD":{"altname":"ETHUSD","wsname":"ETH/USD","base":"XETH","quote":"ZUSD","status":"online"},
	"XDGUSDT":{"altname":"XDGUSDT","wsname":"XDG/USDT","base":"XXDG","quote":"USDT","status":"online"},
	"TRXUSDT":{"altname":"TRXUSDT","wsname":"TRX/USDT","base":"TRX","quote":"USDT","status":"online"},
	"XLTCZEUR":{"altname":"LTCEUR","wsname":"LTC/EUR","base":"XLTC","quote":"ZEUR","status":"online"},
	"BROKEN":{"altname":"BROKEN","wsname":"","base":"X","quote":"Y","status":"online"}
}}`

const krakenTickerResp = `{"error":[],"result":{
	"XBTUSDT":{"c":["50000.0","0.1"]},
	"XETHZUSD":{"c":["2500.00","1"]},
	"XDGUSDT":{"c":["0.125","100"]},
	"TRXUSDT":{"c":["0","1"]},
	"XLTCZEUR":{"c":["80","1"]},
	"UNKNOWNPAIR":{"c":["1","1"]}
}}`

func newTestLogger() logger.Logger {
	return logger.New("test", mxlogger.Config{Level: "fatal"})
}

type krakenServer struct {
	srv         *httptest.Server
	pairsCalls  atomic.Int32
	tickerCalls atomic.Int32
	tickerResp  string
}

func newKrakenServer(t *testing.T) *krakenServer {
	t.Helper()
	ks := &krakenServer{tickerResp: krakenTickerResp}
	ks.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/0/public/AssetPairs":
			ks.pairsCalls.Add(1)
			_, _ = w.Write([]byte(krakenAssetPairsResp))
		case "/0/public/Ticker":
			ks.tickerCalls.Add(1)
			require.Empty(t, r.URL.RawQuery, "all tickers are requested at once")
			_, _ = w.Write([]byte(ks.tickerResp))
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
		}
	}))
	t.Cleanup(ks.srv.Close)
	return ks
}

func collectRates(t *testing.T, f exrate.IFetcher, cf exrate.CurrencyFilter) (map[string]string, error) {
	t.Helper()
	out := make(chan exrate.ExRate, 100)
	err := f.Fetch(t.Context(), cf, out)
	close(out)

	rates := map[string]string{}
	for r := range out {
		require.Equal(t, "kraken", r.Source)
		rates[r.From+"/"+r.To] = r.Value
	}
	return rates, err
}

func TestKrakenFetcher(t *testing.T) {
	ks := newKrakenServer(t)
	f := exrate.NewKrakenFetcher(ks.srv.URL, nil, &http.Client{Timeout: 5 * time.Second}, newTestLogger())
	require.Equal(t, "kraken", f.Source())

	cf := exrate.NewCurrencyFilter(exrate.CurrencyList{
		{Code: "BTC"}, {Code: "ETH"}, {Code: "DOGE"}, {Code: "TRX"},
		{Code: "USDT", IsStable: true}, {Code: "USD", IsStable: true},
	})

	rates, err := collectRates(t, f, cf)
	require.NoError(t, err)

	require.Equal(t, map[string]string{
		"BTC/USDT":  "50000",
		"USDT/BTC":  "0.00002",
		"ETH/USD":   "2500",
		"USD/ETH":   "0.0004",
		"DOGE/USDT": "0.125",
		"USDT/DOGE": "8",
	}, rates, "XBT/XDG are mapped to BTC/DOGE, zero price and unrequested pairs are skipped")
}

func TestKrakenFetcher_CachesAssetPairs(t *testing.T) {
	ks := newKrakenServer(t)
	f := exrate.NewKrakenFetcher(ks.srv.URL, nil, &http.Client{Timeout: 5 * time.Second}, newTestLogger())
	cf := exrate.NewCurrencyFilter(exrate.CurrencyList{{Code: "BTC"}, {Code: "USDT", IsStable: true}})

	for range 3 {
		_, err := collectRates(t, f, cf)
		require.NoError(t, err)
	}
	require.EqualValues(t, 1, ks.pairsCalls.Load())
	require.EqualValues(t, 3, ks.tickerCalls.Load())
}

func TestKrakenFetcher_Errors(t *testing.T) {
	cf := exrate.NewCurrencyFilter(exrate.CurrencyList{{Code: "BTC"}, {Code: "USDT", IsStable: true}})

	t.Run("api error", func(t *testing.T) {
		ks := newKrakenServer(t)
		ks.tickerResp = `{"error":["EService:Unavailable"]}`
		f := exrate.NewKrakenFetcher(ks.srv.URL, nil, &http.Client{Timeout: 5 * time.Second}, newTestLogger())

		_, err := collectRates(t, f, cf)
		require.ErrorContains(t, err, "EService:Unavailable")
	})

	t.Run("empty tickers", func(t *testing.T) {
		ks := newKrakenServer(t)
		ks.tickerResp = `{"error":[],"result":{}}`
		f := exrate.NewKrakenFetcher(ks.srv.URL, nil, &http.Client{Timeout: 5 * time.Second}, newTestLogger())

		_, err := collectRates(t, f, cf)
		require.ErrorContains(t, err, "empty response")
	})

	t.Run("falls back to proxy", func(t *testing.T) {
		ks := newKrakenServer(t)
		// direct requests go to a closed port, the proxy forwards to the fake Kraken
		proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp, err := http.Get(ks.srv.URL + r.URL.Path) //nolint:noctx
			require.NoError(t, err)
			defer resp.Body.Close()
			w.WriteHeader(resp.StatusCode)
			_, _ = io.Copy(w, resp.Body)
		}))
		t.Cleanup(proxy.Close)

		f := exrate.NewKrakenFetcher("http://127.0.0.1:1", []string{proxy.URL}, &http.Client{Timeout: 5 * time.Second}, newTestLogger())
		rates, err := collectRates(t, f, cf)
		require.NoError(t, err)
		require.Equal(t, "50000", rates["BTC/USDT"])
	})
}

// Runs against the real API; enable with KRAKEN_LIVE_TESTS=1.
func TestKrakenFetcher_Live(t *testing.T) {
	if os.Getenv("KRAKEN_LIVE_TESTS") != "1" {
		t.Skip("set KRAKEN_LIVE_TESTS=1 to run")
	}

	f := exrate.NewKrakenFetcher("https://api.kraken.com", nil, &http.Client{Timeout: 15 * time.Second}, newTestLogger())
	cf := exrate.NewCurrencyFilter(exrate.CurrencyList{
		{Code: "BTC"}, {Code: "ETH"}, {Code: "DOGE"}, {Code: "TRX"}, {Code: "LTC"},
		{Code: "USDT", IsStable: true}, {Code: "USDC", IsStable: true},
	})

	rates, err := collectRates(t, f, cf)
	require.NoError(t, err)
	for _, pair := range []string{"BTC/USDT", "ETH/USDT", "DOGE/USDT"} {
		v, ok := rates[pair]
		require.True(t, ok, "missing %s", pair)
		require.True(t, decimal.RequireFromString(v).IsPositive())
	}
	t.Logf("got %d rates: BTC/USDT=%s DOGE/USDT=%s", len(rates), rates["BTC/USDT"], rates["DOGE/USDT"])
}
