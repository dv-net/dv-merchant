package kraken_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/responses"
)

func readForm(t *testing.T, r *http.Request) url.Values {
	t.Helper()
	require.NoError(t, r.ParseForm())
	return r.PostForm
}

func TestSpot_AddOrder(t *testing.T) {
	var userRef int32 = 42
	client := newTestClient(t, "spot-add", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/0/private/AddOrder", r.URL.Path)
		require.NotEmpty(t, r.Header.Get("API-Sign"))

		form := readForm(t, r)
		require.Equal(t, "XBTUSDT", form.Get("pair"))
		require.Equal(t, "sell", form.Get("type"))
		require.Equal(t, "market", form.Get("ordertype"))
		require.Equal(t, "0.0125", form.Get("volume"))
		require.Equal(t, "6d1b345e282140e2ad834ecb18a06876", form.Get("cl_ord_id"))
		require.Equal(t, "42", form.Get("userref"))
		require.Equal(t, "fciq", form.Get("oflags"))
		require.Empty(t, form.Get("price"))
		require.Empty(t, form.Get("validate"))
		require.NotEmpty(t, form.Get("nonce"))

		_, _ = w.Write([]byte(`{"error":[],"result":{
			"descr":{"order":"sell 0.01250000 XBTUSDT @ market"},
			"txid":["OUF4EM-FRGI2-MQMWZD"]
		}}`))
	})

	res, err := client.Spot().AddOrder(t.Context(), &requests.AddOrderRequest{
		Pair:      "XBTUSDT",
		Type:      kraken.OrderSideSell,
		OrderType: kraken.OrderTypeMarket,
		Volume:    decimal.RequireFromString("0.0125"),
		ClOrdID:   "6d1b345e282140e2ad834ecb18a06876",
		UserRef:   &userRef,
		OFlags:    []string{requests.OFlagFeeInQuote},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"OUF4EM-FRGI2-MQMWZD"}, res.TxID)
	require.Equal(t, "sell 0.01250000 XBTUSDT @ market", res.Descr.Order)
}

func TestSpot_AddOrder_LimitAndValidate(t *testing.T) {
	client := newTestClient(t, "spot-validate", func(w http.ResponseWriter, r *http.Request) {
		form := readForm(t, r)
		require.Equal(t, "limit", form.Get("ordertype"))
		require.Equal(t, "37500.5", form.Get("price"))
		require.Equal(t, "true", form.Get("validate"))
		require.Equal(t, "IOC", form.Get("timeinforce"))

		_, _ = w.Write([]byte(`{"error":[],"result":{"descr":{"order":"buy 1.25000000 XBTUSD @ limit 37500.5"}}}`))
	})

	res, err := client.Spot().AddOrder(t.Context(), &requests.AddOrderRequest{
		Pair:        "XBTUSD",
		Type:        kraken.OrderSideBuy,
		OrderType:   kraken.OrderTypeLimit,
		Volume:      decimal.RequireFromString("1.25"),
		Price:       decimal.RequireFromString("37500.5"),
		TimeInForce: "IOC",
		Validate:    true,
	})
	require.NoError(t, err)
	require.Empty(t, res.TxID)
}

func TestSpot_AddOrder_MissingTxID(t *testing.T) {
	client := newTestClient(t, "spot-no-txid", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"error":[],"result":{"descr":{"order":"buy 1 XBTUSD @ market"}}}`))
	})

	_, err := client.Spot().AddOrder(t.Context(), &requests.AddOrderRequest{
		Pair: "XBTUSD", Type: kraken.OrderSideBuy, OrderType: kraken.OrderTypeMarket, Volume: decimal.NewFromInt(1),
	})
	require.ErrorContains(t, err, "without txid")
}

func TestSpot_AddOrder_InvalidRequest(t *testing.T) {
	client := newTestClient(t, "spot-invalid", func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("request must not be sent")
	})

	valid := requests.AddOrderRequest{
		Pair: "XBTUSD", Type: kraken.OrderSideBuy, OrderType: kraken.OrderTypeMarket, Volume: decimal.NewFromInt(1),
	}

	cases := map[string]func(r *requests.AddOrderRequest){
		"no pair":             func(r *requests.AddOrderRequest) { r.Pair = "" },
		"bad side":            func(r *requests.AddOrderRequest) { r.Type = "BUY" },
		"no order type":       func(r *requests.AddOrderRequest) { r.OrderType = "" },
		"zero volume":         func(r *requests.AddOrderRequest) { r.Volume = decimal.Zero },
		"negative volume":     func(r *requests.AddOrderRequest) { r.Volume = decimal.NewFromInt(-1) },
		"limit without price": func(r *requests.AddOrderRequest) { r.OrderType = kraken.OrderTypeLimit },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			req := valid
			mutate(&req)
			_, err := client.Spot().AddOrder(t.Context(), &req)
			require.ErrorIs(t, err, kraken.ErrInvalidOrderRequest)
		})
	}

	_, err := client.Spot().AddOrder(t.Context(), nil)
	require.ErrorIs(t, err, kraken.ErrInvalidOrderRequest)
}

func TestSpot_AddOrder_ErrorMapping(t *testing.T) {
	cases := map[string]error{
		"EOrder:Insufficient funds":           exchangeclient.ErrInsufficientBalance,
		"EOrder:Orders minimum not met":       exchangeclient.ErrMinOrderValue,
		"EOrder:Cost minimum not met":         exchangeclient.ErrMinOrderValue,
		"EService:Market in cancel_only mode": exchangeclient.ErrSymbolTradingHalted,
		"EOrder:Rate limit exceeded":          exchangeclient.ErrRateLimited,
		"EGeneral:Permission denied":          exchangeclient.ErrIncorrectAPIPermissions,
	}
	for krakenErr, want := range cases {
		t.Run(krakenErr, func(t *testing.T) {
			client := newTestClient(t, "spot-err-"+krakenErr, func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(`{"error":["` + krakenErr + `"]}`))
			})
			_, err := client.Spot().AddOrder(t.Context(), &requests.AddOrderRequest{
				Pair: "XBTUSD", Type: kraken.OrderSideBuy, OrderType: kraken.OrderTypeMarket, Volume: decimal.NewFromInt(1),
			})
			require.ErrorIs(t, err, want)
		})
	}
}

func TestSpot_QueryOrders(t *testing.T) {
	client := newTestClient(t, "spot-query", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/0/private/QueryOrders", r.URL.Path)
		form := readForm(t, r)
		require.Equal(t, "OBCMZD-JIEE7-77TH3F,OMMDB2-FSB6Z-7W3HPO", form.Get("txid"))
		require.Equal(t, "true", form.Get("trades"))

		_, _ = w.Write([]byte(`{"error":[],"result":{
			"OBCMZD-JIEE7-77TH3F":{
				"refid":null,"userref":0,"cl_ord_id":"6d1b345e282140e2ad834ecb18a06876",
				"status":"closed","reason":null,
				"opentm":1688665496.7808,"closetm":1688665499.1922,"starttm":0,"expiretm":0,
				"descr":{"pair":"XBTUSDT","type":"sell","ordertype":"market","price":"0","price2":"0",
					"leverage":"none","order":"sell 0.01250000 XBTUSDT @ market","close":""},
				"vol":"0.01250000","vol_exec":"0.01250000","cost":"376.2","fee":"0.97812",
				"price":"30096.0","stopprice":"0.00000","limitprice":"0.00000",
				"misc":"","oflags":"fciq","trades":["TZX2WP-XSEOP-FP7WYR"]
			},
			"OMMDB2-FSB6Z-7W3HPO":{
				"refid":null,"userref":null,"status":"canceled","reason":"User requested",
				"opentm":1688148493.7708,"closetm":1688148610.0482,"starttm":0,"expiretm":0,
				"descr":{"pair":"XBTUSD","type":"buy","ordertype":"limit","price":"27500.0","price2":"0",
					"leverage":"none","order":"buy 1.25 XBTUSD @ limit 27500.0","close":""},
				"vol":"1.25000000","vol_exec":"0.00000000","cost":"0.00000","fee":"0.00000",
				"price":"0.00000","stopprice":"0.00000","limitprice":"0.00000","misc":"","oflags":"fciq"
			}
		}}`))
	})

	res, err := client.Spot().QueryOrders(t.Context(), &requests.QueryOrdersRequest{
		TxIDs:  []string{"OBCMZD-JIEE7-77TH3F", "OMMDB2-FSB6Z-7W3HPO"},
		Trades: true,
	})
	require.NoError(t, err)
	require.Len(t, res, 2)

	filled := res["OBCMZD-JIEE7-77TH3F"]
	require.Equal(t, responses.OrderStatusClosed, filled.Status)
	require.True(t, filled.IsFilled())
	require.True(t, filled.IsFinal())
	require.Equal(t, "6d1b345e282140e2ad834ecb18a06876", filled.ClOrdID)
	require.Equal(t, "XBTUSDT", filled.Descr.Pair)
	require.Equal(t, kraken.OrderSideSell, filled.Descr.Type)
	require.True(t, decimal.RequireFromString("0.0125").Equal(filled.VolExec))
	require.True(t, decimal.RequireFromString("376.2").Equal(filled.Cost))
	require.True(t, decimal.RequireFromString("0.97812").Equal(filled.Fee))
	require.True(t, decimal.RequireFromString("30096").Equal(filled.Price))
	require.Equal(t, []string{"TZX2WP-XSEOP-FP7WYR"}, filled.Trades)
	require.Nil(t, filled.Reason)

	canceled := res["OMMDB2-FSB6Z-7W3HPO"]
	require.False(t, canceled.IsFilled())
	require.True(t, canceled.IsFinal())
	require.NotNil(t, canceled.Reason)
	require.Equal(t, "User requested", *canceled.Reason)
	require.Nil(t, canceled.UserRef)
}

func TestSpot_QueryOrders_InvalidRequest(t *testing.T) {
	client := newTestClient(t, "spot-query-invalid", func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("request must not be sent")
	})

	_, err := client.Spot().QueryOrders(t.Context(), &requests.QueryOrdersRequest{})
	require.ErrorIs(t, err, kraken.ErrInvalidOrderRequest)

	_, err = client.Spot().QueryOrders(t.Context(), &requests.QueryOrdersRequest{TxIDs: make([]string, 51)})
	require.ErrorIs(t, err, kraken.ErrInvalidOrderRequest)
}

func TestSpot_QueryOrders_UnknownOrder(t *testing.T) {
	client := newTestClient(t, "spot-query-unknown", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"error":["EOrder:Unknown order"]}`))
	})

	_, err := client.Spot().QueryOrders(t.Context(), &requests.QueryOrdersRequest{TxIDs: []string{"NOPE"}})
	require.ErrorIs(t, err, kraken.ErrUnknownOrder)
}
