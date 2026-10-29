package kraken

import (
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/responses"
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

// Checks every kraken row of the seed against a real account:
//
//	KRAKEN_API_KEY=... KRAKEN_SECRET_KEY=... go test -run TestSeeds_Live -v ./internal/service/exchange/kraken/
func TestSeeds_Live(t *testing.T) {
	apiKey, secretKey := os.Getenv("KRAKEN_API_KEY"), os.Getenv("KRAKEN_SECRET_KEY")
	if apiKey == "" || secretKey == "" {
		t.Skip("set KRAKEN_API_KEY and KRAKEN_SECRET_KEY to run")
	}

	bURL, err := url.Parse("https://api.kraken.com")
	require.NoError(t, err)
	l := logger.New("test", mxlogger.Config{Level: "fatal"})
	svc, err := NewService(l, apiKey, secretKey, bURL, nil, memory.NewStore(), nil, nil)
	require.NoError(t, err)

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
