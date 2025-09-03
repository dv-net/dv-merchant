package clients

import (
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bitget"

	"github.com/ulule/limiter/v3"
)

type IBitgetSpot interface {
	Account() IBitgetAccount
	Trade() IBitgetTrade
	Market() IBitgetMarket
}

var _ IBitgetSpot = (*SpotClient)(nil)

func NewSpotClient(opt *ClientOptions, store limiter.Store, signer bitget.ISigner) *SpotClient {
	return &SpotClient{
		accountClient: NewAccountClient(opt, store, signer),
		tradeClient:   NewTradeClient(opt, store, signer),
		marketClient:  NewMarketClient(opt, store, signer),
	}
}

type SpotClient struct {
	accountClient IBitgetAccount
	tradeClient   IBitgetTrade
	marketClient  IBitgetMarket
}

func (o *SpotClient) Account() IBitgetAccount { return o.accountClient }
func (o *SpotClient) Trade() IBitgetTrade     { return o.tradeClient }
func (o *SpotClient) Market() IBitgetMarket   { return o.marketClient }
