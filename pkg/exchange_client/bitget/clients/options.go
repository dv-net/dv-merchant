package clients

import (
	"net/http"
	"net/url"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget"
	"github.com/dv-net/dv-merchant/pkg/logger"
)

type ClientOptions struct {
	AccessKey  string
	SecretKey  string
	PassPhrase string
	BaseURL    *url.URL
	Public     bool
}

type ClientOption func(*BaseClient)

func WithSigner(signer bitget.ISigner) ClientOption {
	return func(c *BaseClient) {
		c.signer = signer
	}
}

// WithBaseHTTPClient sets a custom HTTP client for all sub-clients. nil keeps the default.
func WithBaseHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *BaseClient) {
		c.httpClient = httpClient
	}
}

func WithBaseLogger(log logger.Logger) ClientOption {
	return func(c *BaseClient) {
		c.log = log
	}
}

type SubClientOption func(*Client)

func WithLogger(log logger.Logger) SubClientOption {
	return func(c *Client) {
		c.log = log
	}
}

// WithHTTPClient sets a custom HTTP client. nil keeps the default.
func WithHTTPClient(httpClient *http.Client) SubClientOption {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}
