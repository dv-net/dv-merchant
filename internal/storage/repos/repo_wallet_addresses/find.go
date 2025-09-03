package repo_wallet_addresses

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"

	"github.com/dv-net/dv-merchant/pkg/dbutils"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/shopspring/decimal"
)

type FindParams struct {
	Amount     *decimal.Decimal
	CurrencyID *string
	StoreIDs   []uuid.UUID
	WalletIDs  []*uuid.UUID
	Blockchain *string
	Address    *string
	storecmn.CommonFindParams
	Rates         []decimal.Decimal
	IDs           []string
	SortByBalance bool
	SortByAmount  bool
	BalanceFrom   *decimal.Decimal
	BalanceTo     *decimal.Decimal
}

type FindRow struct {
	models.WalletAddress
	CurrencyCode string          `db:"currency_code" json:"currency_code"`
	AmountUSD    decimal.Decimal `db:"amount_usd" json:"amount_usd"`
}

func (s *CustomQuerier) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithFullPagination[*FindRow], error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder()

	storeIDs := make([]interface{}, len(params.StoreIDs))
	for i, id := range params.StoreIDs {
		storeIDs[i] = id
	}

	countSb.Select("COUNT(wallet_addresses.id)").
		From("wallet_addresses").
		JoinWithOption("LEFT", "wallets", "wallets.id = wallet_addresses.wallet_id").
		JoinWithOption("LEFT", "currencies", "wallet_addresses.currency_id = currencies.id").
		JoinWithOption("LEFT", "rate", "currencies.id = rate.currency_id").
		Where(countSb.In("wallets.store_id", storeIDs...))

	rateCte := sqlbuilder.PostgreSQL.NewCTEBuilder().
		With(
			sqlbuilder.CTEQuery("rate", "currency_id", "exchange_rate").As(
				sqlbuilder.Buildf(
					"SELECT unnest(%s::text[]) AS currency_id, unnest(%s::decimal[]) AS exchange_rate",
					params.IDs, params.Rates,
				),
			),
		)

	sb.With(rateCte).
		Select("wallet_addresses.*", "currencies.code AS currency_code", "(wallet_addresses.amount * rate.exchange_rate)::decimal AS amount_usd").
		From("wallet_addresses").
		JoinWithOption("LEFT", "wallets", "wallets.id = wallet_addresses.wallet_id").
		JoinWithOption("LEFT", "currencies", "wallet_addresses.currency_id = currencies.id").
		JoinWithOption("LEFT", "rate", "currencies.id = rate.currency_id").
		Where(sb.In("wallets.store_id", storeIDs...))

	if params.Blockchain != nil {
		sb.Where(sb.Equal("wallet_addresses.blockchain", *params.Blockchain))
		countSb.Where(countSb.Equal("wallet_addresses.blockchain", *params.Blockchain))
	}

	if params.Address != nil {
		sb.Where(sb.Or(
			sb.Equal("wallet_addresses.address", *params.Address),
			sb.Equal("wallet_addresses.address", strings.ToLower(*params.Address)),
		))
		countSb.Where(countSb.Or(
			countSb.Equal("wallet_addresses.address", *params.Address),
			countSb.Equal("wallet_addresses.address", strings.ToLower(*params.Address)),
		))
	}

	if len(params.WalletIDs) > 0 {
		walletIDs := make([]interface{}, len(params.WalletIDs))
		for i, id := range params.WalletIDs {
			walletIDs[i] = id
		}
		sb.Where(sb.In("wallet_addresses.wallet_id", walletIDs...))
		countSb.Where(countSb.In("wallet_addresses.wallet_id", walletIDs...))
	}

	if params.CurrencyID != nil {
		sb.Where(sb.Equal("wallet_addresses.currency_id", *params.CurrencyID))
		countSb.Where(countSb.Equal("wallet_addresses.currency_id", *params.CurrencyID)) // Fix here
	}

	if params.Amount != nil {
		sb.Where(sb.GreaterEqualThan("wallet_addresses.amount", *params.Amount))
		countSb.Where(countSb.GreaterEqualThan("wallet_addresses.amount", *params.Amount)) // Fix here
	}

	if params.BalanceTo != nil {
		sb.Where(sb.LessEqualThan("(wallet_addresses.amount * rate.exchange_rate)", *params.BalanceTo))
		countSb.Where(countSb.LessEqualThan("(wallet_addresses.amount * rate.exchange_rate)", *params.BalanceTo))
	}

	if params.BalanceFrom != nil {
		sb.Where(sb.GreaterEqualThan("(wallet_addresses.amount * rate.exchange_rate)", *params.BalanceFrom))           // Corrected balance filter
		countSb.Where(countSb.GreaterEqualThan("(wallet_addresses.amount * rate.exchange_rate)", *params.BalanceFrom)) // Corrected balance filter
	}

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(500))
	if err != nil {
		return nil, err
	}

	if params.OrderBy == "" {
		params.OrderBy = "amount_usd"
	}

	if params.SortByAmount {
		params.OrderBy = "amount"
	}

	// https://github.com/huandu/go-sqlbuilder/issues/24#issuecomment-496102500
	if !params.IsAscOrdering {
		params.OrderBy += " DESC"
	}
	sb.OrderBy(params.OrderBy)

	sb.OrderBy("wallet_addresses.id")
	if !params.IsAscOrdering {
		sb.Desc()
	}

	sb.Limit(int(limit))   // #nosec
	sb.Offset(int(offset)) // #nosec

	var items []*FindRow

	selectSQL, selectArgs := sb.Build()

	if err := pgxscan.Select(ctx, s.psql, &items, selectSQL, selectArgs...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	var totalCnt uint32
	pagingSQL, args := countSb.With(rateCte).Build()
	if err := pgxscan.Get(ctx, s.psql, &totalCnt, pagingSQL, args...); err != nil {
		return nil, fmt.Errorf("select count: %w", err)
	}

	var currPage uint32 = 1
	if params.Page != nil {
		currPage = *params.Page
	}

	var pagesCnt uint64 = 1
	if params.PageSize != nil {
		pagesCnt = uint64(math.Ceil(float64(totalCnt) / float64(*params.PageSize)))
	}

	return &storecmn.FindResponseWithFullPagination[*FindRow]{
		Items: items,
		Pagination: storecmn.FullPagingData{
			Total:    uint64(totalCnt),
			PageSize: uint64(limit),
			Page:     uint64(currPage),
			LastPage: pagesCnt,
		},
	}, nil
}
