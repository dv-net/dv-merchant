package utils

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewHTTPClient_ReusesTransport(t *testing.T) {
	cfg := NetConfig{ProxyURL: "socks5://127.0.0.1:1080"}

	c1, err := NewHTTPClient(cfg, time.Second)
	require.NoError(t, err)
	c2, err := NewHTTPClient(cfg, time.Second)
	require.NoError(t, err)

	require.Same(t, c1.Transport, c2.Transport)
}

func TestNewHTTPClient_Interface(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("interface names below are Linux-specific")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	t.Run("bound to loopback", func(t *testing.T) {
		c, err := NewHTTPClient(NetConfig{Interface: "lo"}, 5*time.Second)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, http.NoBody)
		require.NoError(t, err)
		resp, err := c.Do(req)
		require.NoError(t, err)
		_ = resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("unknown interface fails", func(t *testing.T) {
		c, err := NewHTTPClient(NetConfig{Interface: "no-such-iface0"}, 5*time.Second)
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, http.NoBody)
		require.NoError(t, err)
		resp, err := c.Do(req)
		if resp != nil {
			_ = resp.Body.Close()
		}
		require.Error(t, err)
	})
}
