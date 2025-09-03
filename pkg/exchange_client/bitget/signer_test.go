package bitget_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget"
)

func TestSigner_SignRequest(t *testing.T) {
	signer := bitget.NewSigner(
		"bg_c98fc8f2a15d906ecb8726d04bfe68a1",
		"564b42b41f6d18b9873500716ffba5ff4c773e5544783874d80198020b39d02b",
		"r200H4UYeRkxxDdWDUCJkGoe",
	)

	t.Run("AllAccountBalance", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.bitget.com/api/v2/account/all-account-balance", nil)
		require.NoError(t, err)
		query := req.URL.Query()
		req.URL.RawQuery = query.Encode()
		newReq, err := signer.SignRequest(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, newReq)
		res, err := http.DefaultClient.Do(newReq)
		t.Log(newReq.URL.RawQuery)
		require.NoError(t, err)
		require.NotNil(t, res)
		defer res.Body.Close()
		resBytes, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		t.Log(string(resBytes))
	})

	t.Run("FundingAssets", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.bitget.com/api/v2/spot/account/assets", nil)
		require.NoError(t, err)
		query := req.URL.Query()
		query.Add("coin", "TRX")
		req.URL.RawQuery = query.Encode()
		newReq, err := signer.SignRequest(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, newReq)
		res, err := http.DefaultClient.Do(newReq)
		t.Log(newReq.URL.String())
		t.Logf("%v", newReq.Header)
		require.NoError(t, err)
		require.NotNil(t, res)
		defer res.Body.Close()
		resBytes, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		t.Log(string(resBytes))
	})
}
