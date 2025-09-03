package repo_transactions

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/shopspring/decimal"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/pkg/dbutils"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
)

type GetByUserParams struct {
	storecmn.CommonFindParams
	UserID        uuid.UUID
	Currencies    []string
	StoreUUIDs    []uuid.UUID
	WalletAddress string
	ToAddresses   string
	FromAddresses string
	Type          *models.TransactionsType
	IsSystem      bool
	Blockchain    *models.Blockchain
	MinAmount     decimal.Decimal
	DateFrom      *time.Time
	DateTo        *time.Time
}

type FindRow struct {
	models.Transaction
	UserEmail      *string `json:"user_email"`
	UntrustedEmail *string `json:"untrusted_email"`
	Currency       struct {
		ID         string             `db:"id" json:"id"`
		Code       string             `db:"code" json:"code"`
		Name       string             `db:"name" json:"name"`
		Blockchain *models.Blockchain `db:"blockchain" json:"blockchain"`
	} `db:"currency_info" json:"currency"`
}

func (s *CustomQuerier) GetByUser(
	ctx context.Context,
	params GetByUserParams,
) (*storecmn.FindResponseWithFullPagination[*FindRow], error) {
	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countSb.Select("COUNT(transactions.id)").
		From("transactions").
		JoinWithOption("INNER", "currencies", "currencies.id = transactions.currency_id").
		Where(countSb.Equal("transactions.user_id", params.UserID.String()))

	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select(
		`transactions.*`,
		`currencies.id as "currency_info.id"`,
		`currencies.code as "currency_info.code"`,
		`currencies.name as "currency_info.name"`,
		`currencies.blockchain as "currency_info.blockchain"`,
		`wallets.email as "user_email"`,
		`wallets.untrusted_email as untrusted_email`,

		// -- hack: partial replacement email value with '*' char
		// `COALESCE(concat(rpad(substring(wallets.email from 1 for 5), char_length(wallets.email) - 6, '*'),
		//      substring(wallets.email from char_length(wallets.email) - 6)), '') as "user_email"`,
	).
		From("transactions").
		JoinWithOption("INNER", "currencies", "currencies.id = transactions.currency_id").
		JoinWithOption("LEFT", "wallets", "wallets.id = transactions.wallet_id").
		Where(sb.Equal("transactions.user_id", params.UserID.String()))

	storeIDs := make([]interface{}, len(params.StoreUUIDs))
	for i, id := range params.StoreUUIDs {
		storeIDs[i] = id
	}
	if len(storeIDs) > 0 {
		sb.Where(sb.In("transactions.store_id", storeIDs...))
		countSb.Where(countSb.In("transactions.store_id", storeIDs...))
	}

	currencies := make([]interface{}, len(params.Currencies))
	for i, id := range params.Currencies {
		currencies[i] = id
	}
	if len(currencies) > 0 {
		sb.Where(sb.In("transactions.currency_id", currencies...))
		countSb.Where(countSb.In("transactions.currency_id", currencies...))
	}

	if params.Type != nil {
		sb.Where(sb.Equal("transactions.type", string(*params.Type)))
		countSb.Where(countSb.Equal("transactions.type", string(*params.Type)))
	}
	if !params.MinAmount.IsZero() {
		sb.Where(sb.GreaterEqualThan("transactions.amount_usd", params.MinAmount))
		countSb.Where(countSb.GreaterEqualThan("transactions.amount_usd", params.MinAmount))
	}

	if params.DateFrom != nil {
		from := params.DateFrom.Unix() * 1000
		sb.Where(
			sb.GreaterEqualThan("transactions.created_at_index", from),
		)
		countSb.Where(
			countSb.GreaterEqualThan("transactions.created_at_index", from),
		)
	}

	if params.DateTo != nil {
		to := params.DateTo.Unix() * 1000
		sb.Where(
			sb.LessEqualThan("transactions.created_at_index", to),
		)
		countSb.Where(
			countSb.LessEqualThan("transactions.created_at_index", to),
		)
	}

	if params.WalletAddress != "" {
		sb.Where(
			sb.Or(
				sb.Equal("transactions.to_address", params.WalletAddress),
				sb.Equal("transactions.from_address", params.WalletAddress),
			),
		)
		countSb.Where(
			countSb.Or(
				countSb.Equal("transactions.to_address", params.WalletAddress),
				countSb.Equal("transactions.from_address", params.WalletAddress),
			),
		)
	}

	if params.ToAddresses != "" {
		sb.Where(sb.Equal("transactions.to_address", params.ToAddresses))
		countSb.Where(countSb.Equal("transactions.to_address", params.ToAddresses))
	}
	if params.FromAddresses != "" {
		sb.Where(sb.Equal("transactions.from_address", params.FromAddresses))
		countSb.Where(countSb.Equal("transactions.from_address", params.FromAddresses))
	}

	if params.Blockchain != nil {
		sb.Where(sb.Equal("transactions.blockchain", params.Blockchain))
		countSb.Where(countSb.Equal("transactions.blockchain", params.Blockchain))
	}

	sb.Where(sb.Equal("transactions.is_system", params.IsSystem))
	countSb.Where(countSb.Equal("transactions.is_system", params.IsSystem))

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(1000))
	if err != nil {
		return nil, err
	}

	if params.OrderBy == "" {
		params.OrderBy = "transactions.created_at_index"
		params.IsAscOrdering = false
	}

	sb.OrderBy(params.OrderBy)
	if !params.IsAscOrdering {
		sb.Desc()
	}
	sb.Limit(int(limit))   // #nosec
	sb.Offset(int(offset)) // #nosec

	items := make([]*FindRow, 0)
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &items, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	var totalCnt uint64
	pagingSQL, args := countSb.Build()
	if err := pgxscan.Get(ctx, s.psql, &totalCnt, pagingSQL, args...); err != nil {
		return nil, fmt.Errorf("select paging query: %w", err)
	}

	var page uint64 = 1
	if params.Page != nil {
		page = uint64(*params.Page)
	}

	var pagesCnt uint64 = 1
	if params.PageSize != nil {
		pagesCnt = uint64(math.Ceil(float64(totalCnt) / float64(*params.PageSize)))
	}
	return &storecmn.FindResponseWithFullPagination[*FindRow]{
		Items: items,
		Pagination: storecmn.FullPagingData{
			Total:    totalCnt,
			PageSize: uint64(limit),
			Page:     page,
			LastPage: pagesCnt,
		},
	}, nil
}
