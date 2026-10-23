package kraken

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/service/currconv"
	"github.com/dv-net/dv-merchant/internal/storage/repos/repo_exchange_chains"
	exchangeclient "github.com/dv-net/dv-merchant/pkg/exchange_client"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/requests"
	"github.com/dv-net/dv-merchant/pkg/exchange_client/kraken/responses"
)

const maxWithdrawalAttempts = 10

// ManualVerificationRequired is matched by exchange_withdrawal to keep transfers enabled
// while Kraken holds the withdrawal for review.
const ManualVerificationRequired = "manual verification required (kraken: onhold)"

// On "insufficient funds" the withdrawal is retried with the amount reduced by this share.
var withdrawalRetryStep = decimal.RequireFromString("0.005")

func (o *Service) GetWithdrawalRules(ctx context.Context, currencies ...string) ([]*models.WithdrawalRulesDTO, error) {
	currEnabled, err := o.storage.ExchangeChains().GetEnabledCurrencies(ctx, models.ExchangeSlugKraken)
	if err != nil {
		return nil, fmt.Errorf("get enabled currencies: %w", err)
	}
	currEnabled = lo.Filter(currEnabled, func(item *repo_exchange_chains.GetEnabledCurrenciesRow, _ int) bool {
		return lo.Contains(currencies, item.ID.String)
	})
	if len(currEnabled) == 0 {
		return []*models.WithdrawalRulesDTO{}, nil
	}

	assets, err := o.assets.get(ctx)
	if err != nil {
		return nil, err
	}
	methods, err := o.exClient.Wallet().GetWithdrawMethods(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get withdraw methods: %w", err)
	}
	addresses, err := o.verifiedWithdrawAddresses(ctx, "", "")
	if err != nil {
		return nil, err
	}

	rules := make([]*models.WithdrawalRulesDTO, 0, len(currEnabled))
	for _, curr := range currEnabled {
		asset, ok := assets.byCode(curr.Ticker)
		if !ok || !asset.CanWithdraw() {
			continue
		}
		method, ok := lo.Find(methods, func(m responses.WithdrawMethod) bool {
			return m.Method == curr.Chain && assets.code(m.Asset) == curr.Ticker
		})
		if !ok {
			continue
		}

		rule := &models.WithdrawalRulesDTO{
			Currency:          curr.Ticker,
			Chain:             curr.Chain,
			MinWithdrawAmount: method.Minimum.String(),
			WithdrawPrecision: strconv.Itoa(asset.Decimals),
			WithdrawFeeType:   models.WithdrawalFeeTypeFixed,
		}

		// Kraken returns the fee only for a saved withdrawal address.
		if saved, ok := lo.Find(addresses, func(a responses.WithdrawAddress) bool {
			return a.Method == curr.Chain && assets.code(a.Asset) == curr.Ticker
		}); ok {
			infoAmount := method.Minimum
			if !infoAmount.IsPositive() {
				infoAmount = decimal.NewFromInt(1)
			}
			info, err := o.exClient.Wallet().GetWithdrawInfo(ctx, &requests.WithdrawInfoRequest{
				Asset:  krakenAsset(curr.Ticker),
				Key:    saved.Key,
				Amount: infoAmount,
			})
			if err != nil {
				return nil, fmt.Errorf("get withdraw info for %s (%s): %w", curr.Ticker, curr.Chain, err)
			}
			rule.Fee = info.Fee.String()
			rule.MaxWithdrawAmount = info.Limit.String()
		}

		minDepositAmount, err := o.convSvc.Convert(ctx, currconv.ConvertDTO{
			Source: models.ExchangeSlugKraken.String(),
			From:   models.CurrencyCodeUSDT,
			To:     curr.Ticker,
			Amount: "1",
		})
		if err != nil {
			return nil, fmt.Errorf("convert minimum deposit amount for %s: %w", curr.Ticker, err)
		}
		rule.MinDepositAmount = minDepositAmount.String()

		rules = append(rules, rule)
	}

	return rules, nil
}

func (o *Service) verifiedWithdrawAddresses(ctx context.Context, asset, method string) (responses.WithdrawAddressesResponse, error) {
	verified := true
	addresses, err := o.exClient.Wallet().GetWithdrawAddresses(ctx, &requests.WithdrawAddressesRequest{
		Asset:    asset,
		Method:   method,
		Verified: &verified,
	})
	if err != nil {
		return nil, fmt.Errorf("get withdraw addresses: %w", err)
	}
	return addresses, nil
}

func (o *Service) CreateWithdrawalOrder(ctx context.Context, args *models.CreateWithdrawalOrderParams) (*models.ExchangeWithdrawalDTO, error) {
	precision := int32(args.WithdrawalPrecision) //nolint:gosec
	amount := args.NativeAmount.RoundDown(precision)

	ticker, err := o.storage.ExchangeChains().GetTickerByCurrencyID(ctx, repo_exchange_chains.GetTickerByCurrencyIDParams{
		CurrencyID: args.Currency,
		Slug:       models.ExchangeSlugKraken,
	})
	if err != nil {
		return nil, fmt.Errorf("get exchange ticker for %s: %w", args.Currency, err)
	}

	if amount.LessThan(args.MinWithdrawal) {
		return nil, exchangeclient.ErrMinWithdrawalBalance
	}

	asset := krakenAsset(ticker)
	addresses, err := o.verifiedWithdrawAddresses(ctx, asset, args.Chain)
	if err != nil {
		return nil, err
	}
	saved, ok := addresses.FindByAddress(args.Address, args.Chain)
	if !ok {
		return nil, fmt.Errorf("address %s is not saved as verified %s (%s) withdrawal address in kraken: %w",
			args.Address, ticker, args.Chain, exchangeclient.ErrWithdrawalAddressNotWhitelisted)
	}

	recordID := ""
	if args.RecordID != nil {
		recordID = args.RecordID.String()
	}
	o.l.Infow(
		"withdrawal request assembled",
		"exchange", models.ExchangeSlugKraken.String(),
		"recordID", recordID,
		"amount", amount.String(),
		"fee", args.Fee.String(),
		"currency", ticker,
		"chain", args.Chain,
		"address", args.Address,
		"key", saved.Key,
	)

	dto := &models.ExchangeWithdrawalDTO{}
	for range maxWithdrawalAttempts {
		res, err := o.exClient.Wallet().Withdraw(ctx, &requests.WithdrawRequest{
			Asset:   asset,
			Key:     saved.Key,
			Amount:  amount,
			Address: saved.Address,
		})
		if err == nil {
			dto.ExternalOrderID = res.RefID
			return dto, nil
		}
		if !errors.Is(err, exchangeclient.ErrInsufficientBalance) {
			return nil, err
		}

		o.l.Errorw("insufficient funds, retrying with reduced amount",
			"error", err,
			"exchange", models.ExchangeSlugKraken.String(),
			"recordID", recordID,
			"current_amount", amount.String(),
		)

		amount = amount.Sub(amount.Mul(withdrawalRetryStep)).RoundDown(precision)
		if amount.LessThan(args.MinWithdrawal) {
			return nil, exchangeclient.ErrMinWithdrawalBalance
		}
		dto.RetryReason = exchangeclient.ErrInsufficientBalance.Error()
	}

	return nil, fmt.Errorf("withdrawal failed after %d attempts: %w", maxWithdrawalAttempts, exchangeclient.ErrInsufficientBalance)
}

// GetWithdrawalByID returns Status already mapped to models.WithdrawalHistoryStatus values;
// a withdrawal held by Kraken for review is reported as recovery.
func (o *Service) GetWithdrawalByID(ctx context.Context, args *models.GetWithdrawalByIDParams) (*models.WithdrawalStatusDTO, error) {
	var refID string
	switch {
	case args.ExternalOrderID != nil && *args.ExternalOrderID != "":
		refID = *args.ExternalOrderID
	case args.ClientOrderID != nil && *args.ClientOrderID != "":
		refID = *args.ClientOrderID
	default:
		return nil, fmt.Errorf("either ClientOrderID or ExternalOrderID must be provided")
	}

	history, err := o.exClient.Wallet().GetWithdrawStatus(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get withdrawal status: %w", err)
	}

	withdrawal, ok := history.FindByRefID(refID)
	if !ok {
		return nil, fmt.Errorf("withdrawal %s not found", refID)
	}

	res := &models.WithdrawalStatusDTO{
		ID:           withdrawal.RefID,
		TxHash:       withdrawal.TxID,
		NativeAmount: withdrawal.Amount,
	}
	switch {
	case withdrawal.IsSuccess():
		res.Status = models.WithdrawalHistoryStatusCompleted.String()
	case withdrawal.IsFailed():
		res.Status = models.WithdrawalHistoryStatusFailed.String()
		res.ErrorMessage = withdrawal.Status
		if withdrawal.StatusProp != "" {
			res.ErrorMessage += " (" + withdrawal.StatusProp + ")"
		}
	case withdrawal.StatusProp == responses.WithdrawStatusPropOnHold:
		res.Status = models.WithdrawalHistoryStatusRecovery.String()
		res.ErrorMessage = ManualVerificationRequired
	default:
		res.Status = models.WithdrawalHistoryStatusInProgress.String()
	}
	return res, nil
}
