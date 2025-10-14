package clients

import (
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

func WithBaseLogger(log logger.Logger) ClientOption {
	return func(c *BaseClient) {
		c.log = log
	}
}

type SubClientOption func(*Client)

func WithLogger(log logger.Logger) SubClientOption {
	return func(c *Client) {
		c.log = log
		c.logEnabled = true
	}
}
