package kucoin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	kucoinrequests "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/requests"
	kucoinresponses "github.com/dv-net/dv-merchant/pkg/exchange_client/kucoin/responses"
	"github.com/dv-net/dv-merchant/pkg/key_value"

	"github.com/ulule/limiter/v3"
)

const (
	getAllSymbols = "/api/v2/symbols"
	getSymbol     = "/api/v2/symbols/%s"
	getTicker     = "/api/v1/market/orderbook/level1"

	// Cache settings for symbol endpoints
	symbolCacheTTL = 5 * time.Minute
)

type IKucoinPublic interface {
	GetSymbol(ctx context.Context, req kucoinrequests.GetSymbol) (*kucoinresponses.GetSymbol, error)
	GetAllSymbols(ctx context.Context, req kucoinrequests.GetAllSymbols) (*kucoinresponses.GetAllSymbols, error)
	GetTicker(ctx context.Context, req kucoinrequests.GetTicker) (*kucoinresponses.GetTicker, error)
}

type Public struct {
	client *Client
	cache  key_value.IKeyValue
}

func NewPublic(clientOpt *ClientOptions, store limiter.Store, cache key_value.IKeyValue, opts ...ClientOption) *Public {
	public := &Public{
		client: NewClient(clientOpt, store, opts...),
		cache:  cache,
	}
	public.initLimiters()
	return public
}

func (o *Public) initLimiters() {
	// KuCoin public endpoints: VIP0 allows 4000 requests per 30 seconds
	// Conservative approach: 100 requests per 30 seconds to stay well under limit
	// This allows for burst traffic while preventing rate limit violations
	publicRate := limiter.Rate{Limit: 100, Period: 30 * time.Second}

	o.client.limiters = map[string]*limiter.Limiter{
		getAllSymbols: limiter.New(o.client.store, publicRate),
		getSymbol:     limiter.New(o.client.store, publicRate), // Weight: 4
		getTicker:     limiter.New(o.client.store, publicRate),
	}
}

// Helper methods for caching
func (o *Public) getCacheKey(endpoint string, params ...string) string {
	key := fmt.Sprintf("kucoin:public:%s", endpoint)
	for _, param := range params {
		key += ":" + param
	}
	return key
}

func (o *Public) getFromCache(ctx context.Context, key string, dest interface{}) bool {
	if o.cache == nil {
		return false
	}

	cached, err := o.cache.Get(ctx, key)
	if err != nil {
		return false
	}

	if json.Unmarshal(cached.Bytes(), dest) == nil {
		return true
	}

	return false
}

func (o *Public) setCache(ctx context.Context, key string, data interface{}, ttl time.Duration) {
	if o.cache == nil {
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	_ = o.cache.Set(ctx, key, jsonData, ttl)
}

func (o *Public) GetAllSymbols(ctx context.Context, req kucoinrequests.GetAllSymbols) (*kucoinresponses.GetAllSymbols, error) {
	response := &kucoinresponses.GetAllSymbols{}

	// Check cache first
	cacheKey := o.getCacheKey("all_symbols")
	if o.getFromCache(ctx, cacheKey, response) {
		return response, nil
	}

	// Not in cache, fetch from API
	p := getAllSymbols
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, false, response, m)
	if err != nil {
		return nil, err
	}

	// Cache the successful response
	o.setCache(ctx, cacheKey, response, symbolCacheTTL)

	return response, nil
}

func (o *Public) GetSymbol(ctx context.Context, req kucoinrequests.GetSymbol) (*kucoinresponses.GetSymbol, error) {
	response := &kucoinresponses.GetSymbol{}

	// Check cache first (use symbol as cache key parameter)
	cacheKey := o.getCacheKey("symbol", req.Symbol)
	if o.getFromCache(ctx, cacheKey, response) {
		return response, nil
	}

	// Not in cache, fetch from API
	p := getSymbol
	m := S2M(req)
	if req.Symbol != "" {
		p = strings.Replace(p, "%s", req.Symbol, 1)
	}
	err := o.client.Do(ctx, http.MethodGet, p, false, response, m)
	if err != nil {
		return nil, err
	}

	// Cache the successful response
	o.setCache(ctx, cacheKey, response, symbolCacheTTL)

	return response, nil
}

func (o *Public) GetTicker(ctx context.Context, req kucoinrequests.GetTicker) (*kucoinresponses.GetTicker, error) {
	response := &kucoinresponses.GetTicker{}
	p := getTicker
	m := S2M(req)
	err := o.client.Do(ctx, http.MethodGet, p, false, response, m)
	if err != nil {
		return nil, err
	}
	return response, nil
}
