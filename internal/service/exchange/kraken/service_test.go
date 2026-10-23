package kraken

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/storage"
	"github.com/dv-net/dv-merchant/internal/storage/repos"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_chains"
	currconvmocks "github.com/dv-net/dv-merchant/mocks/currconv"
	chainsmocks "github.com/dv-net/dv-merchant/mocks/repo_exchange_chains"
	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	"github.com/dv-net/dv-merchant/pkg/logger"
	mxlogger "github.com/dv-net/mx/logger"
)

const assetsJSON = `{"error":[],"result":{
	"XXBT":{"aclass":"currency","altname":"XBT","decimals":10,"display_decimals":5,"status":"enabled"},
	"USDT":{"aclass":"currency","altname":"USDT","decimals":8,"display_decimals":4,"status":"enabled"},
	"XETH":{"aclass":"currency","altname":"ETH","decimals":10,"display_decimals":5,"status":"enabled"},
	"ZUSD":{"aclass":"currency","altname":"USD","decimals":4,"display_decimals":2,"status":"enabled"},
	"XXDG":{"aclass":"currency","altname":"XDG","decimals":8,"display_decimals":2,"status":"deposit_only"}
}}`

const pairXBTUSDT = `"XBTUSDT":{"altname":"XBTUSDT","wsname":"XBT/USDT","base":"XXBT","quote":"USDT",
	"pair_decimals":1,"cost_decimals":5,"lot_decimals":8,"ordermin":"0.00005","costmin":"0.5","tick_size":"0.1","status":"online"}`

type call struct {
	path string
	form url.Values
}

// fakeKraken answers by endpoint path and records every request.
type fakeKraken struct {
	t        *testing.T
	mu       sync.Mutex
	handlers map[string]func(form url.Values) string
	calls    []call
}

func (f *fakeKraken) on(path string, h func(form url.Values) string) {
	f.handlers[path] = h
}

func (f *fakeKraken) reply(path, body string) {
	f.on(path, func(url.Values) string { return body })
}

func (f *fakeKraken) callsTo(path string) []call {
	f.mu.Lock()
	defer f.mu.Unlock()
	var res []call
	for _, c := range f.calls {
		if c.path == path {
			res = append(res, c)
		}
	}
	return res
}

func (f *fakeKraken) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	require.NoError(f.t, r.ParseForm())
	form := r.PostForm
	if r.Method == http.MethodGet {
		form = r.URL.Query()
	}

	f.mu.Lock()
	f.calls = append(f.calls, call{path: r.URL.Path, form: form})
	h, ok := f.handlers[r.URL.Path]
	f.mu.Unlock()

	if !ok {
		f.t.Errorf("unexpected request to %s", r.URL.Path)
		_, _ = w.Write([]byte(`{"error":["EGeneral:Unexpected"]}`))
		return
	}
	_, _ = w.Write([]byte(h(form)))
}

// fakeStorage serves only ExchangeChains; any other storage call panics.
type fakeStorage struct {
	storage.IStorage
	chains repo_exchange_chains.Querier
}

func (f *fakeStorage) ExchangeChains(...repos.Option) repo_exchange_chains.Querier {
	return f.chains
}

type testEnv struct {
	svc    *Service
	kraken *fakeKraken
	chains *chainsmocks.ExchangeChainsQuerier
	conv   *currconvmocks.ICurrencyConvertor
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	fk := &fakeKraken{t: t, handlers: map[string]func(url.Values) string{}}
	fk.reply("/0/public/Assets", assetsJSON)

	srv := httptest.NewServer(fk)
	t.Cleanup(srv.Close)
	bURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	chains := chainsmocks.NewExchangeChainsQuerier(t)
	conv := currconvmocks.NewICurrencyConvertor(t)

	secret := base64.StdEncoding.EncodeToString([]byte("secret"))
	l := logger.New("test", mxlogger.Config{Level: "fatal"})
	svc, err := NewService(l, "key-"+t.Name(), secret, bURL, &fakeStorage{chains: chains}, nil, conv)
	require.NoError(t, err)

	return &testEnv{svc: svc, kraken: fk, chains: chains, conv: conv}
}

func TestAssetsIndex(t *testing.T) {
	env := newTestEnv(t)
	idx, err := env.svc.assets.get(t.Context())
	require.NoError(t, err)

	require.Equal(t, "BTC", idx.code("XXBT"))
	require.Equal(t, "BTC", idx.code("XBT"))
	require.Equal(t, "DOGE", idx.code("XXDG"))
	require.Equal(t, "USD", idx.code("ZUSD"))
	require.Equal(t, "USDT", idx.code("USDT"))
	require.Equal(t, "UNKNOWN", idx.code("UNKNOWN"))

	asset, ok := idx.byCode("BTC")
	require.True(t, ok)
	require.Equal(t, 10, asset.Decimals)

	require.Equal(t, "XBT", krakenAsset("BTC"))
	require.Equal(t, "XDG", krakenAsset("DOGE"))
	require.Equal(t, "USDT", krakenAsset("USDT"))

	// second call is served from cache
	_, err = env.svc.assets.get(t.Context())
	require.NoError(t, err)
	require.Len(t, env.kraken.callsTo("/0/public/Assets"), 1)
}

const balanceExJSON = `{"error":[],"result":{
	"XXBT":{"balance":"0.5","hold_trade":"0.1"},
	"USDT":{"balance":"100","hold_trade":"0"},
	"USDT.F":{"balance":"50","hold_trade":"0"},
	"XETH":{"balance":"0","hold_trade":"0"},
	"ZUSD":{"balance":"10","hold_trade":"0"}
}}`

func TestGetCurrencyBalance(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/private/BalanceEx", balanceExJSON)

	btc, err := env.svc.GetCurrencyBalance(t.Context(), "BTC")
	require.NoError(t, err)
	require.True(t, decimal.RequireFromString("0.4").Equal(*btc))

	usdt, err := env.svc.GetCurrencyBalance(t.Context(), "USDT")
	require.NoError(t, err)
	require.True(t, decimal.NewFromInt(100).Equal(*usdt), "earn balance USDT.F must not be counted")

	missing, err := env.svc.GetCurrencyBalance(t.Context(), "TRX")
	require.NoError(t, err)
	require.True(t, missing.IsZero())
}

func TestGetAccountBalance(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/private/BalanceEx", balanceExJSON)

	env.chains.EXPECT().GetCurrencyIDByTicker(mock.Anything, "BTC").Return("BTC.Bitcoin", nil)
	env.chains.EXPECT().GetCurrencyIDByTicker(mock.Anything, "USDT").Return("USDT.Tron", nil)
	env.chains.EXPECT().GetCurrencyIDByTicker(mock.Anything, "USD").Return("", pgx.ErrNoRows)

	env.conv.EXPECT().Convert(mock.Anything, currconv.ConvertDTO{Source: "kraken", From: "BTC", To: "USDT", Amount: "0.4"}).
		Return(decimal.RequireFromString("40000.123456"), nil)
	env.conv.EXPECT().Convert(mock.Anything, currconv.ConvertDTO{Source: "kraken", From: "USDT", To: "USDT", Amount: "100"}).
		Return(decimal.NewFromInt(100), nil)

	balances, err := env.svc.GetAccountBalance(t.Context())
	require.NoError(t, err)
	require.Len(t, balances, 2)

	byCurrency := map[string]*models.AccountBalanceDTO{}
	for _, b := range balances {
		byCurrency[b.Currency] = b
	}
	require.True(t, decimal.RequireFromString("40000.1235").Equal(byCurrency["BTC.Bitcoin"].AmountUSD))
	require.True(t, decimal.NewFromInt(100).Equal(byCurrency["USDT.Tron"].Amount))
}

func TestGetExchangeSymbols(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/public/AssetPairs", `{"error":[],"result":{`+pairXBTUSDT+`,
		"XXBTZUSD":{"altname":"XBTUSD","base":"XXBT","quote":"ZUSD","status":"online"},
		"XETHXXBT":{"altname":"ETHXBT","base":"XETH","quote":"XXBT","status":"cancel_only"}
	}}`)

	symbols, err := env.svc.GetExchangeSymbols(t.Context())
	require.NoError(t, err)
	require.Len(t, symbols, 2, "fiat and halted pairs are skipped")

	for _, s := range symbols {
		require.Equal(t, "XBTUSDT", s.Symbol)
		require.Equal(t, "BTC", s.BaseSymbol)
		require.Equal(t, "USDT", s.QuoteSymbol)
	}
	require.ElementsMatch(t, []string{"BTC/USDT", "USDT/BTC"}, []string{symbols[0].DisplayName, symbols[1].DisplayName})
}

func TestGetOrderRule(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/public/AssetPairs", `{"error":[],"result":{`+pairXBTUSDT+`}}`)

	rule, err := env.svc.GetOrderRule(t.Context(), "XBTUSDT")
	require.NoError(t, err)
	require.Equal(t, &models.OrderRulesDTO{
		Symbol:          "XBTUSDT",
		State:           "online",
		BaseCurrency:    "BTC",
		QuoteCurrency:   "USDT",
		MinOrderAmount:  "0.00005",
		MinOrderValue:   "0.5",
		AmountPrecision: 8,
		PricePrecision:  1,
		ValuePrecision:  5,
	}, rule)
}

func TestGetOrderRule_Halted(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/public/AssetPairs", `{"error":[],"result":{"XBTUSDT":{"altname":"XBTUSDT","base":"XXBT","quote":"USDT","status":"cancel_only"}}}`)

	_, err := env.svc.GetOrderRule(t.Context(), "XBTUSDT")
	require.ErrorIs(t, err, exchangeclient.ErrSymbolTradingHalted)
}

func TestGetOrderRule_CanonicalKey(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/public/AssetPairs", `{"error":[],"result":{"XXBTZUSD":{"altname":"XBTUSD","base":"XXBT","quote":"ZUSD","status":"online","ordermin":"0.0001","costmin":"0.5"}}}`)

	rule, err := env.svc.GetOrderRule(t.Context(), "XBTUSD")
	require.NoError(t, err)
	require.Equal(t, "XBTUSD", rule.Symbol)
	require.Equal(t, "USD", rule.QuoteCurrency)
}

var btcUsdtRule = &models.OrderRulesDTO{
	Symbol: "XBTUSDT", BaseCurrency: "BTC", QuoteCurrency: "USDT",
	MinOrderAmount: "0.00005", MinOrderValue: "0.5", AmountPrecision: 8, ValuePrecision: 5,
}

const tickerXBTUSDT = `{"error":[],"result":{"XBTUSDT":{"a":["50000.0","1","1"],"b":["49900.0","1","1"],"c":["49950.0","0.1"]}}}`

func TestCreateSpotOrder_Sell(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/public/Ticker", tickerXBTUSDT)
	env.kraken.reply("/0/private/BalanceEx", `{"error":[],"result":{"XXBT":{"balance":"0.123456789","hold_trade":"0"}}}`)
	env.kraken.reply("/0/private/AddOrder", `{"error":[],"result":{"descr":{"order":"sell"},"txid":["OTX-1"]}}`)

	order, err := env.svc.CreateSpotOrder(t.Context(), "", "", "sell", "XBTUSDT", nil, btcUsdtRule)
	require.NoError(t, err)
	require.Equal(t, "OTX-1", order.ExchangeOrderID)
	require.Len(t, order.ClientOrderID, 32)
	require.True(t, decimal.RequireFromString("0.12345678").Equal(order.Amount))

	calls := env.kraken.callsTo("/0/private/AddOrder")
	require.Len(t, calls, 1)
	form := calls[0].form
	require.Equal(t, "XBTUSDT", form.Get("pair"))
	require.Equal(t, "sell", form.Get("type"))
	require.Equal(t, "market", form.Get("ordertype"))
	require.Equal(t, "0.12345678", form.Get("volume"))
	require.Equal(t, order.ClientOrderID, form.Get("cl_ord_id"))
	require.Empty(t, form.Get("oflags"))
}

func TestCreateSpotOrder_Buy(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/public/Ticker", tickerXBTUSDT)
	env.kraken.reply("/0/private/BalanceEx", `{"error":[],"result":{"USDT":{"balance":"1000","hold_trade":"0"}}}`)
	env.kraken.reply("/0/private/AddOrder", `{"error":[],"result":{"descr":{"order":"buy"},"txid":["OTX-2"]}}`)

	order, err := env.svc.CreateSpotOrder(t.Context(), "", "", "buy", "XBTUSDT", nil, btcUsdtRule)
	require.NoError(t, err)

	// 1000 * 0.99 / 50000 = 0.0198
	form := env.kraken.callsTo("/0/private/AddOrder")[0].form
	require.Equal(t, "buy", form.Get("type"))
	require.Equal(t, "0.0198", form.Get("volume"))
	require.Equal(t, "fcib", form.Get("oflags"))
	require.True(t, decimal.NewFromInt(990).Equal(order.Amount))
}

func TestCreateSpotOrder_InsufficientBalance(t *testing.T) {
	cases := map[string]struct {
		side    string
		balance string
	}{
		"sell below ordermin": {"sell", `"XXBT":{"balance":"0.00001","hold_trade":"0"}`},
		"sell all on hold":    {"sell", `"XXBT":{"balance":"1","hold_trade":"1"}`},
		"buy below costmin":   {"buy", `"USDT":{"balance":"0.4","hold_trade":"0"}`},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			env.kraken.reply("/0/public/Ticker", tickerXBTUSDT)
			env.kraken.reply("/0/private/BalanceEx", `{"error":[],"result":{`+tc.balance+`}}`)

			_, err := env.svc.CreateSpotOrder(t.Context(), "", "", tc.side, "XBTUSDT", nil, btcUsdtRule)
			require.ErrorIs(t, err, exchangeclient.ErrInsufficientBalance)
			require.Empty(t, env.kraken.callsTo("/0/private/AddOrder"))
		})
	}
}

func TestCreateSpotOrder_RateLimitedSkipsOrder(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/public/Ticker", `{"error":["EAPI:Rate limit exceeded"]}`)

	_, err := env.svc.CreateSpotOrder(t.Context(), "", "", "sell", "XBTUSDT", nil, btcUsdtRule)
	require.ErrorIs(t, err, exchangeclient.ErrSkipOrder)
}

func TestGetOrderDetails(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/public/AssetPairs", `{"error":[],"result":{`+pairXBTUSDT+`}}`)
	env.kraken.on("/0/private/QueryOrders", func(form url.Values) string {
		require.Equal(t, "OTX-1", form.Get("txid"))
		return `{"error":[],"result":{"OTX-1":{"status":"closed","vol":"0.1","vol_exec":"0.1","cost":"4990.5","fee":"12.4","price":"49905","descr":{"pair":"XBTUSDT","type":"sell","ordertype":"market"}}}}`
	})

	res, err := env.svc.GetOrderDetails(t.Context(), &models.GetOrderByIDParams{
		InstrumentID:    new("XBTUSDT"),
		ExternalOrderID: new("OTX-1"),
	})
	require.NoError(t, err)
	require.Equal(t, models.ExchangeOrderStatusCompleted, res.State)
	require.True(t, decimal.RequireFromString("0.1").Equal(res.Amount))
	require.True(t, decimal.RequireFromString("4990.5").Equal(res.AmountUSD))
}

func TestGetOrderDetails_States(t *testing.T) {
	cases := map[string]struct {
		order  string
		state  models.ExchangeOrderStatus
		reason string
	}{
		"open":     {`{"status":"open","vol_exec":"0"}`, models.ExchangeOrderStatusInProgress, ""},
		"pending":  {`{"status":"pending","vol_exec":"0"}`, models.ExchangeOrderStatusInProgress, ""},
		"canceled": {`{"status":"canceled","vol_exec":"0","reason":"Insufficient funds"}`, models.ExchangeOrderStatusFailed, "Insufficient funds"},
		"expired":  {`{"status":"expired","vol_exec":"0"}`, models.ExchangeOrderStatusFailed, ""},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			env := newTestEnv(t)
			env.kraken.reply("/0/public/AssetPairs", `{"error":[],"result":{`+pairXBTUSDT+`}}`)
			env.kraken.reply("/0/private/QueryOrders", `{"error":[],"result":{"OTX-1":`+tc.order+`}}`)

			res, err := env.svc.GetOrderDetails(t.Context(), &models.GetOrderByIDParams{
				InstrumentID:    new("XBTUSDT"),
				ExternalOrderID: new("OTX-1"),
			})
			require.NoError(t, err)
			require.Equal(t, tc.state, res.State)
			require.Equal(t, tc.reason, res.FailReason)
		})
	}
}

func TestGetOrderDetails_RequiresTxID(t *testing.T) {
	env := newTestEnv(t)

	_, err := env.svc.GetOrderDetails(t.Context(), &models.GetOrderByIDParams{
		InstrumentID:  new("XBTUSDT"),
		ClientOrderID: new("client-id"),
	})
	require.ErrorContains(t, err, "txid")
}

func TestGetDepositAddresses_GeneratesWhenEmpty(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/private/DepositMethods", `{"error":[],"result":[{"method":"Tether USD (TRC20)","limit":false,"gen-address":true}]}`)
	env.kraken.on("/0/private/DepositAddresses", func(form url.Values) string {
		require.Equal(t, "USDT", form.Get("asset"))
		require.Equal(t, "Tether USD (TRC20)", form.Get("method"))
		if form.Get("new") == "true" {
			return `{"error":[],"result":[{"address":"TNewAddress","expiretm":"0","new":true}]}`
		}
		return `{"error":[],"result":[]}`
	})
	env.chains.EXPECT().GetCurrencyIDByParams(mock.Anything, repo_exchange_chains.GetCurrencyIDByParamsParams{
		Slug: models.ExchangeSlugKraken, Chain: "Tether USD (TRC20)", Ticker: "USDT",
	}).Return("USDT.Tron", nil)

	res, err := env.svc.GetDepositAddresses(t.Context(), "USDT", "Tether USD (TRC20)")
	require.NoError(t, err)
	require.Equal(t, []*models.DepositAddressDTO{{
		Address:          "TNewAddress",
		Currency:         "USDT.Tron",
		Chain:            "Tether USD (TRC20)",
		InternalCurrency: "USDT",
		AddressType:      models.DepositAddress,
	}}, res)
	require.Len(t, env.kraken.callsTo("/0/private/DepositAddresses"), 2)
}

func TestGetDepositAddresses_UsesKrakenAssetNameAndTag(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.on("/0/private/DepositMethods", func(form url.Values) string {
		require.Equal(t, "XBT", form.Get("asset"))
		return `{"error":[],"result":[{"method":"Bitcoin","limit":false,"gen-address":true}]}`
	})
	env.kraken.on("/0/private/DepositAddresses", func(form url.Values) string {
		require.Equal(t, "XBT", form.Get("asset"))
		return `{"error":[],"result":[
			{"address":"bc1qexisting","expiretm":"0","tag":"t1"},
			{"address":"bc1qexisting","expiretm":"0","tag":"t1"}
		]}`
	})
	env.chains.EXPECT().GetCurrencyIDByParams(mock.Anything, mock.Anything).Return("BTC.Bitcoin", nil)

	res, err := env.svc.GetDepositAddresses(t.Context(), "BTC", "Bitcoin")
	require.NoError(t, err)
	require.Len(t, res, 1, "duplicates are removed")
	require.Equal(t, "t1", res[0].PaymentTag)
}

func TestGetDepositAddresses_TooManyAddresses(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/private/DepositMethods", `{"error":[],"result":[{"method":"Tether USD (TRC20)","limit":false,"gen-address":true}]}`)
	attempt := 0
	env.kraken.on("/0/private/DepositAddresses", func(form url.Values) string {
		attempt++
		switch {
		case attempt == 1:
			return `{"error":[],"result":[]}`
		case form.Get("new") == "true":
			return `{"error":["EFunding:Too many addresses"]}`
		default:
			return `{"error":[],"result":[{"address":"TOld","expiretm":"0"}]}`
		}
	})
	env.chains.EXPECT().GetCurrencyIDByParams(mock.Anything, mock.Anything).Return("USDT.Tron", nil)

	res, err := env.svc.GetDepositAddresses(t.Context(), "USDT", "Tether USD (TRC20)")
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "TOld", res[0].Address)
}

func TestGetDepositAddresses_UnifiedMethod(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/private/DepositMethods", `{"error":[],"result":[
		{"method":"USDT - Ethereum (Unified)","limit":false,"gen-address":true},
		{"method":"Tether USD (TRC20)","limit":false,"gen-address":true}
	]}`)
	env.kraken.on("/0/private/DepositAddresses", func(form url.Values) string {
		require.Equal(t, "USDT - Ethereum (Unified)", form.Get("method"))
		return `{"error":[],"result":[{"address":"0xabc","expiretm":"0"}]}`
	})
	env.chains.EXPECT().GetCurrencyIDByParams(mock.Anything, repo_exchange_chains.GetCurrencyIDByParamsParams{
		Slug: models.ExchangeSlugKraken, Chain: "Tether USD (ERC20)", Ticker: "USDT",
	}).Return("USDT.Ethereum", nil)

	res, err := env.svc.GetDepositAddresses(t.Context(), "USDT", "Tether USD (ERC20)")
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "0xabc", res[0].Address)
	require.Equal(t, "Tether USD (ERC20)", res[0].Chain, "chain stays the withdrawal method from exchange_chains")
}

func TestGetDepositAddresses_ArbitrumOneUnified(t *testing.T) {
	env := newTestEnv(t)
	// method list as returned by a live Kraken account
	env.kraken.reply("/0/private/DepositMethods", `{"error":[],"result":[
		{"method":"USDT - Ethereum (Unified)","limit":false,"gen-address":true},
		{"method":"Tether USD (TRC20)","limit":false,"gen-address":true},
		{"method":"USDT - Polygon (Unified)","limit":false,"gen-address":true},
		{"method":"USDT - Arbitrum One (Unified)","limit":false,"gen-address":true}
	]}`)
	env.kraken.on("/0/private/DepositAddresses", func(form url.Values) string {
		require.Equal(t, "USDT - Arbitrum One (Unified)", form.Get("method"))
		return `{"error":[],"result":[{"address":"0xarb","expiretm":"0"}]}`
	})
	env.chains.EXPECT().GetCurrencyIDByParams(mock.Anything, mock.Anything).Return("USDT.Arbitrum", nil)

	res, err := env.svc.GetDepositAddresses(t.Context(), "USDT", "Arbitrum One Network")
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.Equal(t, "0xarb", res[0].Address)
}

func TestGetDepositAddresses_MethodNotFound(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/private/DepositMethods", `{"error":[],"result":[
		{"method":"USDT - Solana (Unified)","limit":false,"gen-address":true},
		{"method":"Tether USD (TRC20)","limit":false,"gen-address":true}
	]}`)

	_, err := env.svc.GetDepositAddresses(t.Context(), "USDT", "Tether USD (BEP20)")
	require.ErrorIs(t, err, ErrDepositMethodNotFound)
	require.ErrorContains(t, err, `"Tether USD (TRC20)"`, "error lists methods available on Kraken")
	require.Empty(t, env.kraken.callsTo("/0/private/DepositAddresses"))
}

func TestGetDepositAddresses_NoGenAddress(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/private/DepositMethods", `{"error":[],"result":[{"method":"Bitcoin","limit":false,"gen-address":false}]}`)
	env.kraken.on("/0/private/DepositAddresses", func(form url.Values) string {
		require.NotEqual(t, "true", form.Get("new"), "address must not be generated when gen-address is false")
		return `{"error":[],"result":[]}`
	})

	res, err := env.svc.GetDepositAddresses(t.Context(), "BTC", "Bitcoin")
	require.NoError(t, err)
	require.Empty(t, res)
	require.Len(t, env.kraken.callsTo("/0/private/DepositAddresses"), 1)
}

const withdrawAddressesJSON = `{"error":[],"result":[
	{"address":"TDest","asset":"USDT","method":"Tether USD (TRC20)","key":"my-tron","verified":true}
]}`

func withdrawalParams() *models.CreateWithdrawalOrderParams {
	recordID := uuid.New()
	return &models.CreateWithdrawalOrderParams{
		RecordID:            &recordID,
		Currency:            "USDT.Tron",
		Chain:               "Tether USD (TRC20)",
		NativeAmount:        decimal.RequireFromString("100.123456789"),
		Address:             "TDest",
		Fee:                 decimal.NewFromInt(1),
		WithdrawalPrecision: 6,
		MinWithdrawal:       decimal.NewFromInt(10),
	}
}

func TestCreateWithdrawalOrder(t *testing.T) {
	env := newTestEnv(t)
	env.chains.EXPECT().GetTickerByCurrencyID(mock.Anything, repo_exchange_chains.GetTickerByCurrencyIDParams{
		CurrencyID: "USDT.Tron", Slug: models.ExchangeSlugKraken,
	}).Return("USDT", nil)
	env.kraken.on("/0/private/WithdrawAddresses", func(form url.Values) string {
		require.Equal(t, "USDT", form.Get("asset"))
		require.Equal(t, "Tether USD (TRC20)", form.Get("method"))
		require.Equal(t, "true", form.Get("verified"))
		return withdrawAddressesJSON
	})
	env.kraken.reply("/0/private/Withdraw", `{"error":[],"result":{"refid":"REF-1"}}`)

	res, err := env.svc.CreateWithdrawalOrder(t.Context(), withdrawalParams())
	require.NoError(t, err)
	require.Equal(t, "REF-1", res.ExternalOrderID)
	require.Empty(t, res.RetryReason)

	form := env.kraken.callsTo("/0/private/Withdraw")[0].form
	require.Equal(t, "USDT", form.Get("asset"))
	require.Equal(t, "my-tron", form.Get("key"))
	require.Equal(t, "100.123456", form.Get("amount"))
	require.Equal(t, "TDest", form.Get("address"))
}

func TestCreateWithdrawalOrder_AddressNotSaved(t *testing.T) {
	env := newTestEnv(t)
	env.chains.EXPECT().GetTickerByCurrencyID(mock.Anything, mock.Anything).Return("USDT", nil)
	env.kraken.reply("/0/private/WithdrawAddresses", withdrawAddressesJSON)

	params := withdrawalParams()
	params.Address = "TOther"

	_, err := env.svc.CreateWithdrawalOrder(t.Context(), params)
	require.ErrorIs(t, err, exchangeclient.ErrWithdrawalAddressNotWhitelisted)
	require.Empty(t, env.kraken.callsTo("/0/private/Withdraw"))
}

func TestCreateWithdrawalOrder_RetriesOnInsufficientFunds(t *testing.T) {
	env := newTestEnv(t)
	env.chains.EXPECT().GetTickerByCurrencyID(mock.Anything, mock.Anything).Return("USDT", nil)
	env.kraken.reply("/0/private/WithdrawAddresses", withdrawAddressesJSON)
	attempt := 0
	env.kraken.on("/0/private/Withdraw", func(url.Values) string {
		attempt++
		if attempt < 3 {
			return `{"error":["EFunding:Insufficient funds"]}`
		}
		return `{"error":[],"result":{"refid":"REF-3"}}`
	})

	res, err := env.svc.CreateWithdrawalOrder(t.Context(), withdrawalParams())
	require.NoError(t, err)
	require.Equal(t, "REF-3", res.ExternalOrderID)
	require.Equal(t, exchangeclient.ErrInsufficientBalance.Error(), res.RetryReason)

	calls := env.kraken.callsTo("/0/private/Withdraw")
	require.Len(t, calls, 3)
	require.Equal(t, "100.123456", calls[0].form.Get("amount"))
	require.Equal(t, "99.622838", calls[1].form.Get("amount"))
	require.Equal(t, "99.124723", calls[2].form.Get("amount"))
}

func TestCreateWithdrawalOrder_BelowMinimum(t *testing.T) {
	env := newTestEnv(t)
	env.chains.EXPECT().GetTickerByCurrencyID(mock.Anything, mock.Anything).Return("USDT", nil)

	params := withdrawalParams()
	params.NativeAmount = decimal.NewFromInt(5)

	_, err := env.svc.CreateWithdrawalOrder(t.Context(), params)
	require.ErrorIs(t, err, exchangeclient.ErrMinWithdrawalBalance)
}

func TestGetWithdrawalByID(t *testing.T) {
	env := newTestEnv(t)
	env.kraken.reply("/0/private/WithdrawStatus", `{"error":[],"result":[
		{"refid":"REF-OK","txid":"0xhash","amount":"99","fee":"1","status":"Success"},
		{"refid":"REF-PENDING","txid":"","amount":"10","status":"Pending"},
		{"refid":"REF-CANCELED","txid":"","amount":"10","status":"Pending","status-prop":"canceled"},
		{"refid":"REF-FAIL","txid":"","amount":"10","status":"Failure"},
		{"refid":"REF-HOLD","txid":"","amount":"10","status":"Pending","status-prop":"onhold"}
	]}`)

	cases := map[string]struct {
		status, errMsg string
	}{
		"REF-OK":       {models.WithdrawalHistoryStatusCompleted.String(), ""},
		"REF-PENDING":  {models.WithdrawalHistoryStatusInProgress.String(), ""},
		"REF-CANCELED": {models.WithdrawalHistoryStatusFailed.String(), "Pending (canceled)"},
		"REF-FAIL":     {models.WithdrawalHistoryStatusFailed.String(), "Failure"},
		"REF-HOLD":     {models.WithdrawalHistoryStatusRecovery.String(), ManualVerificationRequired},
	}
	for refID, tc := range cases {
		t.Run(refID, func(t *testing.T) {
			res, err := env.svc.GetWithdrawalByID(t.Context(), &models.GetWithdrawalByIDParams{ClientOrderID: new(refID)})
			require.NoError(t, err)
			require.Equal(t, refID, res.ID)
			require.Equal(t, tc.status, res.Status)
			require.Equal(t, tc.errMsg, res.ErrorMessage)
		})
	}

	ok, err := env.svc.GetWithdrawalByID(t.Context(), &models.GetWithdrawalByIDParams{ExternalOrderID: new("REF-OK")})
	require.NoError(t, err)
	require.Equal(t, "0xhash", ok.TxHash)
	require.True(t, decimal.NewFromInt(99).Equal(ok.NativeAmount))

	_, err = env.svc.GetWithdrawalByID(t.Context(), &models.GetWithdrawalByIDParams{ExternalOrderID: new("NOPE")})
	require.ErrorContains(t, err, "not found")
}

func TestGetWithdrawalRules(t *testing.T) {
	env := newTestEnv(t)
	env.chains.EXPECT().GetEnabledCurrencies(mock.Anything, models.ExchangeSlugKraken).Return([]*repo_exchange_chains.GetEnabledCurrenciesRow{
		{ID: pgtype.Text{String: "USDT.Tron", Valid: true}, Ticker: "USDT", Chain: "Tether USD (TRC20)"},
		{ID: pgtype.Text{String: "BTC.Bitcoin", Valid: true}, Ticker: "BTC", Chain: "Bitcoin"},
		{ID: pgtype.Text{String: "DOGE.Dogecoin", Valid: true}, Ticker: "DOGE", Chain: "Dogecoin"},
		{ID: pgtype.Text{String: "ETH.Ethereum", Valid: true}, Ticker: "ETH", Chain: "Ether"},
	}, nil)
	env.kraken.reply("/0/private/WithdrawMethods", `{"error":[],"result":[
		{"asset":"USDT","method":"Tether USD (TRC20)","network":"Tron","minimum":"5"},
		{"asset":"XXBT","method":"Bitcoin","network":"Bitcoin","minimum":"0.0005"},
		{"asset":"XXDG","method":"Dogecoin","network":"Dogecoin","minimum":"50"}
	]}`)
	env.kraken.reply("/0/private/WithdrawAddresses", withdrawAddressesJSON)
	env.kraken.on("/0/private/WithdrawInfo", func(form url.Values) string {
		require.Equal(t, "USDT", form.Get("asset"))
		require.Equal(t, "my-tron", form.Get("key"))
		require.Equal(t, "5", form.Get("amount"))
		return `{"error":[],"result":{"method":"Tether USD (TRC20)","limit":"2500","amount":"2","fee":"3"}}`
	})
	env.conv.EXPECT().Convert(mock.Anything, mock.MatchedBy(func(d currconv.ConvertDTO) bool { return d.To == "USDT" })).
		Return(decimal.NewFromInt(1), nil)
	env.conv.EXPECT().Convert(mock.Anything, mock.MatchedBy(func(d currconv.ConvertDTO) bool { return d.To == "BTC" })).
		Return(decimal.RequireFromString("0.00002"), nil)

	rules, err := env.svc.GetWithdrawalRules(t.Context(), "USDT.Tron", "BTC.Bitcoin", "DOGE.Dogecoin")
	require.NoError(t, err)
	require.Len(t, rules, 2, "DOGE is deposit_only, ETH is not requested")

	byCurrency := map[string]*models.WithdrawalRulesDTO{}
	for _, r := range rules {
		byCurrency[r.Currency] = r
	}

	usdt := byCurrency["USDT"]
	require.Equal(t, "5", usdt.MinWithdrawAmount)
	require.Equal(t, "3", usdt.Fee)
	require.Equal(t, "2500", usdt.MaxWithdrawAmount)
	require.Equal(t, "8", usdt.WithdrawPrecision)
	require.Equal(t, "1", usdt.MinDepositAmount)

	btc := byCurrency["BTC"]
	require.Equal(t, "0.0005", btc.MinWithdrawAmount)
	require.Empty(t, btc.Fee, "no saved address, so the fee is unknown")
	require.Equal(t, "10", btc.WithdrawPrecision)
}
