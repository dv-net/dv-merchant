package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseProxyURL(t *testing.T) {
	valid := []string{
		"http://203.0.113.10:3128",
		"https://proxy.example.com:443",
		"socks5://user:pass@203.0.113.10:1080",
		"socks5h://203.0.113.10:1080/",
		" SOCKS5://203.0.113.10:1080 ",
	}
	for _, raw := range valid {
		_, err := ParseProxyURL(raw)
		require.NoError(t, err, raw)
	}

	invalid := []string{
		"",
		"203.0.113.10:1080",               // no scheme
		"ftp://203.0.113.10:21",           // unsupported scheme
		"socks5://203.0.113.10",           // no port
		"http://:3128",                    // no host
		"http://203.0.113.10:3128/path",   // path
		"http://203.0.113.10:3128?a=b",    // query
		"socks5://user:secret@bad host:1", // parse error
	}
	for _, raw := range invalid {
		_, err := ParseProxyURL(raw)
		require.ErrorIs(t, err, ErrInvalidProxyURL, raw)
		require.NotContains(t, err.Error(), "secret")
	}
}
