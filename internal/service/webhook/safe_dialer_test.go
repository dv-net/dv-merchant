package webhook

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dv-net/dv-merchant/internal/config"
)

func TestIsForbiddenIP(t *testing.T) {
	forbidden := []string{
		"127.0.0.1", "10.1.2.3", "172.16.0.1", "192.168.1.1", "169.254.169.254",
		"0.0.0.0", "100.64.0.1", "255.255.255.255", "224.0.0.1",
		"::1", "::", "fe80::1", "fc00::1", "::ffff:127.0.0.1", "::ffff:169.254.169.254",
		"64:ff9b::7f00:1",
	}
	for _, s := range forbidden {
		require.True(t, isForbiddenIP(netip.MustParseAddr(s)), s)
	}

	allowed := []string{"1.1.1.1", "8.8.8.8", "93.184.216.34", "2606:4700:4700::1111"}
	for _, s := range allowed {
		require.False(t, isForbiddenIP(netip.MustParseAddr(s)), s)
	}
}

func TestSendWebhookBlocksPrivateDestinations(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"success":true}`))
	}))
	defer srv.Close()

	newClient := func(allow bool) *http.Client {
		s := New(config.WebHook{AllowPrivateNetworks: allow}, nil, nil).(*service) //nolint:forcetypeassert
		return s.httpClient
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL, http.NoBody)
	require.NoError(t, err)

	_, err = newClient(false).Do(req) //nolint:bodyclose
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrForbiddenDestination), err.Error())

	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL, http.NoBody)
	require.NoError(t, err)
	resp, err := newClient(true).Do(req)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
