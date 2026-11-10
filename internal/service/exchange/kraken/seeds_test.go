package kraken

import (
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/responses"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/utils"
	"github.com/dv-net/dv-merchant/pkg/logger"
	mxlogger "github.com/dv-net/mx/logger"
)

const currencyChainsSeed = "../../../../sql/postgres/seeds/currency_chains.up.sql"

type seedRow struct {
	currencyID, ticker, chain string
}

func krakenSeedRows(t *testing.T) []seedRow {
	t.Helper()
	data, err := os.ReadFile(currencyChainsSeed)
	require.NoError(t, err)

	re := regexp.MustCompile(`\('kraken',\s*'([^']+)',\s*'([^']+)',\s*'([^']+)'\)`)
	matches := re.FindAllStringSubmatch(string(data), -1)
	rows := make([]seedRow, 0, len(matches))
	for _, m := range matches {
		rows = append(rows, seedRow{currencyID: m[1], ticker: m[2], chain: m[3]})
	}
	require.NotEmpty(t, rows, "no kraken rows in %s", currencyChainsSeed)
	return rows
}

// EVM rows need a Unified deposit method, see unifiedDepositMethods.
func TestSeeds_EVMRowsHaveUnifiedDepositMethod(t *testing.T) {
	for _, row := range krakenSeedRows(t) {
		network := row.currencyID[strings.Index(row.currencyID, ".")+1:]
		if network != "Ethereum" && network != "Polygon" && network != "Arbitrum" {
			continue
		}
		_, ok := unifiedDepositMethods[row.ticker+"|"+row.chain]
		require.True(t, ok, "add %q to unifiedDepositMethods for %s", row.ticker+"|"+row.chain, row.currencyID)
	}
}

// liveService connects to a real Kraken account from env, optionally through KRAKEN_PROXY_URL
// (http://user:pass@host:port or socks5://host:port).
func liveService(t *testing.T) *Service {
	t.Helper()
	apiKey, secretKey := os.Getenv("KRAKEN_API_KEY"), os.Getenv("KRAKEN_SECRET_KEY")
	if apiKey == "" || secretKey == "" {
		t.Skip("set KRAKEN_API_KEY and KRAKEN_SECRET_KEY to run")
	}

	var httpClient *http.Client
	if proxyURL := os.Getenv("KRAKEN_PROXY_URL"); proxyURL != "" {
		var err error
		httpClient, err = utils.NewHTTPClient(utils.NetConfig{ProxyURL: proxyURL}, 30*time.Second)
		require.NoError(t, err)
	}

	bURL, err := url.Parse("https://api.kraken.com")
	require.NoError(t, err)
	l := logger.New("test", mxlogger.Config{Level: "fatal"})
	svc, err := NewService(l, apiKey, secretKey, bURL, nil, memory.NewStore(), nil, httpClient)
	require.NoError(t, err)
	return svc
}

// Checks every kraken row of the seed against a real account:
//
//	KRAKEN_API_KEY=... KRAKEN_SECRET_KEY=... [KRAKEN_PROXY_URL=...] go test -run TestSeeds_Live -v ./internal/service/exchange/kraken/
func TestSeeds_Live(t *testing.T) {
	svc := liveService(t)

	withdrawMethods, err := svc.exClient.Wallet().GetWithdrawMethods(t.Context(), nil)
	require.NoError(t, err)
	assets, err := svc.assets.get(t.Context())
	require.NoError(t, err)

	for _, row := range krakenSeedRows(t) {
		t.Run(row.currencyID, func(t *testing.T) {
			forAsset := lo.Filter(withdrawMethods, func(m responses.WithdrawMethod, _ int) bool {
				return assets.code(m.Asset) == row.ticker
			})
			_, ok := lo.Find(forAsset, func(m responses.WithdrawMethod) bool { return m.Method == row.chain })
			require.True(t, ok, "withdraw method %q not found for %s, available: %q", row.chain, row.ticker,
				lo.Map(forAsset, func(m responses.WithdrawMethod, _ int) string { return m.Method + " / " + m.Network }))

			deposit, err := svc.resolveDepositMethod(t.Context(), row.ticker, row.chain)
			require.NoError(t, err)
			t.Logf("%s: withdraw %q, deposit %q", row.currencyID, row.chain, deposit.Method)
		})
	}
}

// Prints Kraken withdraw and deposit methods for every merchant currency,
// used to fill kraken rows of the seed:
//
//	KRAKEN_API_KEY=... KRAKEN_SECRET_KEY=... [KRAKEN_PROXY_URL=...] go test -run TestSeeds_DumpMethods -v ./internal/service/exchange/kraken/
func TestSeeds_DumpMethods(t *testing.T) {
	svc := liveService(t)

	withdrawMethods, err := svc.exClient.Wallet().GetWithdrawMethods(t.Context(), nil)
	require.NoError(t, err)
	assets, err := svc.assets.get(t.Context())
	require.NoError(t, err)

	tickers := []string{
		"ARB", "BCH", "BNB", "BTC", "CAKE", "DAI", "DOGE", "ENA", "ETH", "LTC", "PEPE", "POL", "PYUSD",
		"SAND", "SHIB", "TRX", "USD1", "USDC", "USDD", "USDE", "USDT", "WLD", "WLFI", "XAUT",
	}
	for _, ticker := range tickers {
		for _, m := range withdrawMethods {
			if assets.code(m.Asset) == ticker {
				t.Logf("%-6s withdraw %q network %q", ticker, m.Method, m.Network)
			}
		}
		deposits, err := svc.exClient.Wallet().GetDepositMethods(t.Context(), &requests.DepositMethodsRequest{Asset: krakenAsset(ticker)})
		if err != nil {
			t.Logf("%-6s deposit error: %v", ticker, err)
			continue
		}
		for _, m := range deposits {
			t.Logf("%-6s deposit  %q", ticker, m.Method)
		}
	}
}
