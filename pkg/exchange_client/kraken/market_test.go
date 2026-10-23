package kraken_test

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken"
)

func TestMarket_GetAssets(t *testing.T) {
	client := newTestClient(t, "", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "/0/public/Assets", r.URL.Path)
		require.Equal(t, "XBT,USDT", r.URL.Query().Get("asset"))
		require.Empty(t, r.Header.Get("API-Sign"))

		_, _ = w.Write([]byte(`{"error":[],"result":{
			"XXBT":{"aclass":"currency","altname":"XBT","decimals":10,"display_decimals":5,"collateral_value":1,"status":"enabled"},
			"USDT":{"aclass":"currency","altname":"USDT","decimals":8,"display_decimals":4,"status":"deposit_only"}
		}}`))
	})

	res, err := client.Market().GetAssets(t.Context(), "XBT", "USDT")
	require.NoError(t, err)
	require.Len(t, res, 2)

	require.Equal(t, "XBT", res["XXBT"].AltName)
	require.Equal(t, 10, res["XXBT"].Decimals)
	require.True(t, res["XXBT"].CanWithdraw())

	require.True(t, res["USDT"].CanDeposit())
	require.False(t, res["USDT"].CanWithdraw())
}

func TestMarket_GetAssetPairs(t *testing.T) {
	client := newTestClient(t, "", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/0/public/AssetPairs", r.URL.Path)
		require.Equal(t, "XBTUSDT", r.URL.Query().Get("pair"))

		_, _ = w.Write([]byte(`{"error":[],"result":{"XBTUSDT":{
			"altname":"XBTUSDT","wsname":"XBT/USDT","aclass_base":"currency","base":"XXBT",
			"aclass_quote":"currency","quote":"USDT","lot":"unit","cost_decimals":5,
			"pair_decimals":1,"lot_decimals":8,"lot_multiplier":1,
			"leverage_buy":[2,3],"leverage_sell":[2,3],
			"fees":[[0,0.4],[10000,0.35]],"fees_maker":[[0,0.25]],"fee_volume_currency":"ZUSD",
			"margin_call":80,"margin_stop":40,
			"ordermin":"0.00005","costmin":"0.5","tick_size":"0.1","status":"online"
		}}}`))
	})

	res, err := client.Market().GetAssetPairs(t.Context(), "XBTUSDT")
	require.NoError(t, err)

	pair, ok := res["XBTUSDT"]
	require.True(t, ok)
	require.Equal(t, "XBT/USDT", pair.WsName)
	require.Equal(t, "XXBT", pair.Base)
	require.Equal(t, "USDT", pair.Quote)
	require.Equal(t, 8, pair.LotDecimals)
	require.Equal(t, 1, pair.PairDecimals)
	require.True(t, decimal.RequireFromString("0.00005").Equal(pair.OrderMin))
	require.True(t, decimal.RequireFromString("0.5").Equal(pair.CostMin))
	require.True(t, decimal.RequireFromString("0.1").Equal(pair.TickSize))
	require.True(t, pair.IsTradable())
}

func TestMarket_GetTicker(t *testing.T) {
	client := newTestClient(t, "", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/0/public/Ticker", r.URL.Path)
		require.Equal(t, "XBTUSD", r.URL.Query().Get("pair"))

		// Kraken answers with the canonical pair name, not the requested one.
		_, _ = w.Write([]byte(`{"error":[],"result":{"XXBTZUSD":{
			"a":["30300.10000","1","1.000"],
			"b":["30300.00000","1","1.000"],
			"c":["30303.20000","0.00067643"],
			"v":["4083.67001100","4412.73601799"],
			"p":["30706.77771","30689.13205"],
			"t":[34619,38907],
			"l":["29868.30000","29868.30000"],
			"h":["31631.00000","31631.00000"],
			"o":"30502.80000"
		}}}`))
	})

	res, err := client.Market().GetTicker(t.Context(), "XBTUSD")
	require.NoError(t, err)

	ticker, ok := res["XXBTZUSD"]
	require.True(t, ok)

	ask, err := ticker.AskPrice()
	require.NoError(t, err)
	require.True(t, decimal.RequireFromString("30300.1").Equal(ask))

	bid, err := ticker.BidPrice()
	require.NoError(t, err)
	require.True(t, decimal.RequireFromString("30300").Equal(bid))

	last, err := ticker.LastPrice()
	require.NoError(t, err)
	require.True(t, decimal.RequireFromString("30303.2").Equal(last))

	require.True(t, decimal.RequireFromString("4412.73601799").Equal(ticker.Volume24h()))
	require.Equal(t, []int64{34619, 38907}, ticker.Trades)
	require.True(t, decimal.RequireFromString("30502.8").Equal(ticker.Open))
}

func TestMarket_NoFilter(t *testing.T) {
	client := newTestClient(t, "", func(w http.ResponseWriter, r *http.Request) {
		require.Empty(t, r.URL.RawQuery)
		_, _ = w.Write([]byte(`{"error":[],"result":{}}`))
	})

	_, err := client.Market().GetTicker(t.Context())
	require.NoError(t, err)
}

func TestMarket_UnknownPair(t *testing.T) {
	client := newTestClient(t, "", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"error":["EQuery:Unknown asset pair"]}`))
	})

	_, err := client.Market().GetAssetPairs(t.Context(), "FOOBAR")
	require.ErrorIs(t, err, kraken.ErrUnknownAssetPair)
}

func TestMarket_UnknownAsset(t *testing.T) {
	client := newTestClient(t, "", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"error":["EQuery:Unknown asset"]}`))
	})

	_, err := client.Market().GetAssets(t.Context(), "FOO")
	require.ErrorIs(t, err, kraken.ErrUnknownAsset)
	require.NotErrorIs(t, err, kraken.ErrUnknownAssetPair)
}

// Runs against the real API; enable with KRAKEN_LIVE_TESTS=1.
func TestMarket_Live(t *testing.T) {
	if os.Getenv("KRAKEN_LIVE_TESTS") != "1" {
		t.Skip("set KRAKEN_LIVE_TESTS=1 to run")
	}

	bURL, err := url.Parse("https://api.kraken.com")
	require.NoError(t, err)

	client, err := kraken.NewBaseClient(&kraken.ClientOptions{BaseURL: bURL}, memory.NewStore())
	require.NoError(t, err)

	assets, err := client.Market().GetAssets(t.Context(), "XBT", "USDT")
	require.NoError(t, err)
	require.Contains(t, assets, "XXBT")

	pairs, err := client.Market().GetAssetPairs(t.Context(), "XBTUSDT")
	require.NoError(t, err)
	require.Contains(t, pairs, "XBTUSDT")
	t.Logf("XBTUSDT: %+v", pairs["XBTUSDT"])

	ticker, err := client.Market().GetTicker(t.Context(), "XBTUSDT")
	require.NoError(t, err)
	last, err := ticker["XBTUSDT"].LastPrice()
	require.NoError(t, err)
	require.True(t, last.IsPositive())
	t.Logf("XBTUSDT last=%s", last)
}
