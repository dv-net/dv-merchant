package kraken

import (
	"context"
	"net/url"
	"strings"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/responses"
)

const (
	getAssetsEndpoint     = "/0/public/Assets"
	getAssetPairsEndpoint = "/0/public/AssetPairs"
	getTickerEndpoint     = "/0/public/Ticker"
)

type IKrakenMarket interface {
	GetAssets(ctx context.Context, assets ...string) (responses.AssetsResponse, error)
	GetAssetPairs(ctx context.Context, pairs ...string) (responses.AssetPairsResponse, error)
	GetTicker(ctx context.Context, pairs ...string) (responses.TickerResponse, error)
}

var _ IKrakenMarket = (*MarketClient)(nil)

type MarketClient struct {
	client *Client
}

func NewMarketClient(client *Client) *MarketClient {
	return &MarketClient{client: client}
}

func (o *MarketClient) GetAssets(ctx context.Context, assets ...string) (responses.AssetsResponse, error) {
	return DoPublic[responses.AssetsResponse](ctx, o.client, getAssetsEndpoint, listParam("asset", assets))
}

func (o *MarketClient) GetAssetPairs(ctx context.Context, pairs ...string) (responses.AssetPairsResponse, error) {
	return DoPublic[responses.AssetPairsResponse](ctx, o.client, getAssetPairsEndpoint, listParam("pair", pairs))
}

func (o *MarketClient) GetTicker(ctx context.Context, pairs ...string) (responses.TickerResponse, error) {
	return DoPublic[responses.TickerResponse](ctx, o.client, getTickerEndpoint, listParam("pair", pairs))
}

func listParam(name string, values []string) url.Values {
	if len(values) == 0 {
		return nil
	}
	return url.Values{name: {strings.Join(values, ",")}}
}
