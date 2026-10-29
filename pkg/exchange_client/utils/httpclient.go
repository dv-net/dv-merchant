package utils

import (
	"net/http"
	"sync"
	"time"
)

type NetConfig struct {
	ProxyURL  string
	Interface string
}

// One transport per config, so connections are reused instead of leaking
// a new pool every time a driver is created.
var (
	transportsMu sync.Mutex
	transports   = make(map[NetConfig]*http.Transport)
)

func NewHTTPClient(cfg NetConfig, timeout time.Duration) (*http.Client, error) {
	transportsMu.Lock()
	defer transportsMu.Unlock()

	t, ok := transports[cfg]
	if !ok {
		var err error
		if t, err = newTransport(cfg); err != nil {
			return nil, err
		}
		transports[cfg] = t
	}

	return &http.Client{Transport: t, Timeout: timeout}, nil
}

// newTransport mirrors http.DefaultTransport settings, but with its own dialer and proxy
// (and no proxy from HTTP_PROXY env when cfg has none).
func newTransport(cfg NetConfig) (*http.Transport, error) {
	t := &http.Transport{
		DialContext:           newDialContext(cfg.Interface),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if cfg.ProxyURL != "" {
		u, err := ParseProxyURL(cfg.ProxyURL)
		if err != nil {
			return nil, err
		}
		t.Proxy = http.ProxyURL(u)
	}
	return t, nil
}
