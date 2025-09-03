package bybit

import (
	"context"
	"net/http"

	"github.com/ulule/limiter/v3"

	"github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/bybit/responses"
)

type IBybitMarket interface {
	GetInstruments(ctx context.Context, req *requests.GetInstrumentsRequest) (*responses.BaseResponse[responses.GetInstrumentsResponse], error)
	GetTickers(ctx context.Context, req *requests.GetTickersRequest) (*responses.BaseResponse[responses.GetTickersResponse], error)
	GetCoinInfo(ctx context.Context, req *requests.GetCoinInfoRequest) (*responses.BaseResponse[responses.GetCoinInfoResponse], error)
}

func NewMarket(opt *ClientOptions, store limiter.Store, opts ...ClientOption) IBybitMarket {
	client := NewClient(opt, store, opts...)
	return &MarketClient{
		client: client,
	}
}

type MarketClient struct {
	client *Client
}

func (o *MarketClient) GetInstruments(ctx context.Context, req *requests.GetInstrumentsRequest) (*responses.BaseResponse[responses.GetInstrumentsResponse], error) {
	endpoint := "/v5/market/instruments-info"

	params := S2M(req)
	var resp responses.BaseResponse[responses.GetInstrumentsResponse]
	err := o.client.Do(ctx, http.MethodGet, endpoint, true, &resp, params)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *MarketClient) GetTickers(ctx context.Context, req *requests.GetTickersRequest) (*responses.BaseResponse[responses.GetTickersResponse], error) {
	endpoint := "/v5/market/tickers"

	params := S2M(req)
	var resp responses.BaseResponse[responses.GetTickersResponse]
	err := o.client.Do(ctx, http.MethodGet, endpoint, false, &resp, params)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (o *MarketClient) GetCoinInfo(ctx context.Context, req *requests.GetCoinInfoRequest) (*responses.BaseResponse[responses.GetCoinInfoResponse], error) {
	endpoint := "/v5/asset/coin/query-info"

	params := S2M(req)
	var resp responses.BaseResponse[responses.GetCoinInfoResponse]
	err := o.client.Do(ctx, http.MethodGet, endpoint, true, &resp, params)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
