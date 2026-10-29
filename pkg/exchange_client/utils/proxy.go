package utils

import (
	"errors"
	"net/url"
	"strings"
)

var ErrInvalidProxyURL = errors.New("invalid proxy url: expected scheme://[user:password@]host:port with scheme http, https, socks5 or socks5h")

var allowedProxySchemes = map[string]struct{}{
	"http":    {},
	"https":   {},
	"socks5":  {},
	"socks5h": {},
}

func ParseProxyURL(raw string) (*url.URL, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, ErrInvalidProxyURL
	}
	if _, ok := allowedProxySchemes[strings.ToLower(u.Scheme)]; !ok {
		return nil, ErrInvalidProxyURL
	}
	if u.Hostname() == "" || u.Port() == "" {
		return nil, ErrInvalidProxyURL
	}
	if (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
		return nil, ErrInvalidProxyURL
	}
	return u, nil
}
