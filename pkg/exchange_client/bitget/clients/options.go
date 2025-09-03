package clients

import (
	"net/url"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget"
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
