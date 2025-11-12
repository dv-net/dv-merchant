package repo_wallets

import (
	"context"
	"fmt"

	"github.com/dv-net/dv-merchant/internal/models"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type ICustomQuerier interface {
	Querier
	SummarizeByCurrency(ctx context.Context, params SummarizeByCurrencyParams) ([]*SummarizeByCurrencyRow, error)
}

type CustomQuerier struct {
	*Queries
	psql DBTX
}

func NewCustom(psql DBTX) *CustomQuerier {
	return &CustomQuerier{
		Queries: New(psql),
		psql:    psql,
	}
}

func (s *CustomQuerier) WithTx(tx pgx.Tx) *CustomQuerier {
	return &CustomQuerier{
		Queries: New(tx),
		psql:    tx,
	}
}

type SummarizeByCurrencyRow struct {
	ID               string             `db:"id" json:"id"`
	Code             string             `db:"code" json:"code"`
	Name             string             `db:"name" json:"name"`
	SortOrder        int64              `db:"sort_order" json:"sort_order"`
	Blockchain       *models.Blockchain `db:"blockchain" json:"blockchain"`
	Count            int64              `db:"count" json:"count"`
	Balance          decimal.Decimal    `db:"balance" json:"balance"`
	CountWithBalance int64              `db:"count_with_balance" json:"count_with_balance"`
	RateSource       string             `db:"rate_source" json:"rate_source"`
	AmountUsd        decimal.Decimal    `db:"amount_usd" json:"amount_usd"`
	IsNative         bool               `db:"is_native" json:"is_native"`
	ContractAddress  string             `db:"contract_address" json:"contract_address"`
}

type SummarizeByCurrencyParams struct {
	UserID     uuid.UUID
	IDs        []string
	Rates      []decimal.Decimal
	MinBalance decimal.Decimal
}

func (s *CustomQuerier) SummarizeByCurrency(ctx context.Context, params SummarizeByCurrencyParams) ([]*SummarizeByCurrencyRow, error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()

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
		Select(
			"currencies.id",
			"currencies.code",
			"currencies.name",
			"currencies.blockchain",
			"currencies.sort_order",
			"currencies.is_native",
			"currencies.contract_address",
			"COUNT(wallet_addresses.id)",
			fmt.Sprintf(
				`SUM(CASE WHEN (wallet_addresses.amount * rate.exchange_rate)::decimal >= %s THEN 1 ELSE 0 END) AS count_with_balance`,
				params.MinBalance,
			),
			fmt.Sprintf(
				`SUM(CASE WHEN (wallet_addresses.amount * rate.exchange_rate)::decimal >= %s THEN wallet_addresses.amount ELSE 0 END)::numeric(90,50) AS balance`,
				params.MinBalance,
			),
			fmt.Sprintf(
				`SUM(CASE WHEN (wallet_addresses.amount * rate.exchange_rate)::decimal >= %s THEN wallet_addresses.amount * rate.exchange_rate ELSE 0 END)::numeric(90,50) AS amount_usd`,
				params.MinBalance,
			),
		).
		From("wallet_addresses").
		JoinWithOption("LEFT", "currencies", "wallet_addresses.currency_id = currencies.id").
		JoinWithOption("LEFT", "wallets", "wallets.id = wallet_addresses.wallet_id").
		JoinWithOption("LEFT", "stores", "stores.id = wallets.store_id").
		JoinWithOption("LEFT", "rate", "currencies.id = rate.currency_id").
		Where(
			sb.Equal("wallet_addresses.user_id", params.UserID),
		)

	sb.GroupBy(
		"currencies.id",
	)

	var items []*SummarizeByCurrencyRow

	selectSQL, args := sb.Build()
	err := pgxscan.Select(ctx, s.psql, &items, selectSQL, args...)
	if err != nil {
		return nil, err
	}

	return items, nil
}
