package kraken

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	kraken "github.com/dv-net/dv-merchant/pkg/exchange_client/kraken"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/responses"
)

const assetsCacheTTL = 10 * time.Minute

// krakenAsset converts our currency code to the name accepted by Kraken API.
func krakenAsset(code string) string {
	return kraken.AssetName(code)
}

// isEarnAsset reports balances like "USDT.F" or "XBT.M" that belong to Kraken Earn.
func isEarnAsset(name string) bool {
	return strings.Contains(name, ".")
}

type assetsIndex struct {
	assets responses.AssetsResponse
	codes  map[string]string // Kraken asset name and alt name -> our code
}

func newAssetsIndex(assets responses.AssetsResponse) *assetsIndex {
	idx := &assetsIndex{
		assets: assets,
		codes:  make(map[string]string, len(assets)*2),
	}
	for name, asset := range assets {
		code := kraken.CurrencyCode(asset.AltName)
		idx.codes[name] = code
		idx.codes[asset.AltName] = code
	}
	return idx
}

// code converts Kraken asset name ("XXBT", "XBT", "ZUSD") to our currency code ("BTC", "USD").
func (i *assetsIndex) code(name string) string {
	if code, ok := i.codes[name]; ok {
		return code
	}
	return kraken.CurrencyCode(name)
}

// byCode finds Kraken asset info by our currency code.
func (i *assetsIndex) byCode(code string) (responses.Asset, bool) {
	for name, asset := range i.assets {
		if i.code(name) == code {
			return asset, true
		}
	}
	return responses.Asset{}, false
}

// Asset list is public and rarely changes, so it is shared by all services of one Kraken host.
type assetsCache struct {
	market kraken.IKrakenMarket
	mu     sync.Mutex
	index  *assetsIndex
	at     time.Time
}

var assetsCaches sync.Map // base url -> *assetsCache

func assetsCacheFor(baseURL string, market kraken.IKrakenMarket) *assetsCache {
	cache, _ := assetsCaches.LoadOrStore(baseURL, &assetsCache{market: market})
	return cache.(*assetsCache) //nolint:forcetypeassert
}

func (c *assetsCache) get(ctx context.Context) (*assetsIndex, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.index != nil && time.Since(c.at) < assetsCacheTTL {
		return c.index, nil
	}

	assets, err := c.market.GetAssets(ctx)
	if err != nil {
		if c.index != nil {
			return c.index, nil
		}
		return nil, fmt.Errorf("get assets: %w", err)
	}

	c.index, c.at = newAssetsIndex(assets), time.Now()
	return c.index, nil
}
