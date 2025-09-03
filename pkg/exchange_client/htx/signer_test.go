package htx

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSigner_SignRequest(t *testing.T) {
	t.Skip()
	signer := &Signer{
		accessKey: "",
		secretKey: []byte(""),
	}
	t.Run("SignRequest", func(t *testing.T) {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.huobi.pro/v1/account/accounts", nil)
		require.NoError(t, err)
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
}
