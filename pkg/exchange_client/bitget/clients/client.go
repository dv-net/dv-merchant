package clients

import (
	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget"
)

var _ IBaseBitgetClient = (*BaseClient)(nil)

type IBaseBitgetClient interface {
	Spot() IBitgetSpot
	Common() IBitgetCommon
	Signer() bitget.ISigner
}

func NewBaseClient(opt *ClientOptions, store limiter.Store, opts ...ClientOption) (*BaseClient, error) {
	c := &BaseClient{
		signer: bitget.NewSigner(opt.AccessKey, opt.SecretKey, opt.PassPhrase),
	}

	for _, opt := range opts {
		opt(c)
	}

	c.commonClient = NewCommonClient(opt, store, c.signer)
	c.spotClient = NewSpotClient(opt, store, c.signer)
	return c, nil
}

type BaseClient struct {
	signer       bitget.ISigner
	spotClient   IBitgetSpot
	commonClient IBitgetCommon
}

func (o *BaseClient) Spot() IBitgetSpot      { return o.spotClient }
func (o *BaseClient) Common() IBitgetCommon  { return o.commonClient }
func (o *BaseClient) Signer() bitget.ISigner { return o.signer }
