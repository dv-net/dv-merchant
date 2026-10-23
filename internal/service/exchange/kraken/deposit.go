package kraken

import (
	"cmp"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/samber/lo"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_chains"
	kraken "github.com/dv-net/dv-merchant/pkg/exchange_client/kraken"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/responses"
)

// Since March 2024 Kraken deposits on EVM networks use "Unified" method names,
// while withdrawal methods kept the old ones. exchange_chains.chain stores the
// withdrawal method, so the deposit method is looked up by our code and that name.
// Candidates are tried in order; names confirmed on a live account go first.
var unifiedDepositMethods = map[string][]string{
	"USDT|Tether USD (ERC20)":            {"USDT - Ethereum (Unified)"},
	"USDT|Tether USD (Polygon)":          {"USDT - Polygon (Unified)"},
	"USDT|Arbitrum One Network":          {"USDT - Arbitrum One (Unified)", "USDT - Arbitrum (Unified)"},
	"USDC|USDC":                          {"USDC - Ethereum (Unified)"},
	"USDC|Polygon - Native":              {"USDC - Polygon (Unified)"},
	"USDC|Arbitrum One Network - Native": {"USDC - Arbitrum One (Unified)", "USDC - Arbitrum (Unified)"},
	"ETH|Ether":                          {"Ether (Hex)", "ETH - Ethereum (Unified)"},
}

var ErrDepositMethodNotFound = errors.New("kraken deposit method not found")

// resolveDepositMethod picks the deposit method for a chain (withdrawal method name):
// the same name if Kraken has it, otherwise the Unified name from the table above.
func (o *Service) resolveDepositMethod(ctx context.Context, currency, chain string) (responses.DepositMethod, error) {
	methods, err := o.exClient.Wallet().GetDepositMethods(ctx, &requests.DepositMethodsRequest{Asset: krakenAsset(currency)})
	if err != nil {
		return responses.DepositMethod{}, fmt.Errorf("get deposit methods for %s: %w", currency, err)
	}

	if m, ok := methods.Find(chain); ok {
		return m, nil
	}
	for _, candidate := range unifiedDepositMethods[currency+"|"+chain] {
		if m, ok := methods.Find(candidate); ok {
			return m, nil
		}
	}

	available := lo.Map(methods, func(m responses.DepositMethod, _ int) string { return m.Method })
	return responses.DepositMethod{}, fmt.Errorf("%w: %s (%s), available: %q", ErrDepositMethodNotFound, currency, chain, available)
}

// GetDepositAddresses expects chain to be a Kraken withdrawal method name
// ("Tether USD (TRC20)"), as stored in exchange_chains.chain.
func (o *Service) GetDepositAddresses(ctx context.Context, currency, chain string) ([]*models.DepositAddressDTO, error) {
	method, err := o.resolveDepositMethod(ctx, currency, chain)
	if err != nil {
		return nil, err
	}

	req := &requests.DepositAddressesRequest{Asset: krakenAsset(currency), Method: method.Method}

	addresses, err := o.exClient.Wallet().GetDepositAddresses(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("get deposit addresses for %s (%s): %w", currency, method.Method, err)
	}

	if len(addresses) == 0 && method.GenAddress {
		req.New = true
		addresses, err = o.exClient.Wallet().GetDepositAddresses(ctx, req)
		if err != nil && !errors.Is(err, kraken.ErrTooManyAddresses) {
			return nil, fmt.Errorf("create deposit address for %s (%s): %w", currency, method.Method, err)
		}
		if errors.Is(err, kraken.ErrTooManyAddresses) {
			req.New = false
			if addresses, err = o.exClient.Wallet().GetDepositAddresses(ctx, req); err != nil {
				return nil, fmt.Errorf("get deposit addresses for %s (%s): %w", currency, method.Method, err)
			}
		}
	}
	if len(addresses) == 0 {
		return nil, nil
	}

	currencyID, err := o.storage.ExchangeChains().GetCurrencyIDByParams(ctx, repo_exchange_chains.GetCurrencyIDByParamsParams{
		Ticker: currency,
		Chain:  chain,
		Slug:   models.ExchangeSlugKraken,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get internal currency id for %s: %w", chain, err)
	}

	unique := lo.UniqBy(addresses, func(item responses.DepositAddress) string {
		return item.Address + ":" + cmp.Or(item.Tag, item.Memo)
	})

	return lo.Map(unique, func(item responses.DepositAddress, _ int) *models.DepositAddressDTO {
		return &models.DepositAddressDTO{
			Address:          item.Address,
			Currency:         currencyID,
			Chain:            chain,
			InternalCurrency: currency,
			AddressType:      models.DepositAddress,
			PaymentTag:       cmp.Or(item.Tag, item.Memo),
		}
	}), nil
}
