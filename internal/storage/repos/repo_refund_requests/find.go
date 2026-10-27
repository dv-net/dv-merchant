package repo_refund_requests

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/dv-net/dv-merchant/internal/models"
	"github.com/dv-net/dv-merchant/internal/storage/storecmn"
	"github.com/dv-net/dv-merchant/pkg/dbutils"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type FindParams struct {
	storecmn.CommonFindParams
	UserID             uuid.UUID
	Statuses           []string
	Blockchain         *models.Blockchain
	TxHash             string // deposit or refund transfer hash
	FromAddress        string // deposit sender
	DestinationAddress string
	DateFrom           *time.Time
	DateTo             *time.Time
}

// FindRow is a refund request with the blocked deposit, its AML check and the refund transfer
type FindRow struct {
	models.RefundRequest

	// blocked deposit
	TransactionID  uuid.UUID           `db:"tx_id"`
	TxHash         string              `db:"tx_hash"`
	TxAmount       decimal.Decimal     `db:"tx_amount"`
	TxAmountUsd    decimal.NullDecimal `db:"tx_amount_usd"`
	TxCurrencyID   string              `db:"tx_currency_id"`
	TxCurrencyCode string              `db:"tx_currency_code"`
	TxBlockchain   models.Blockchain   `db:"tx_blockchain"`
	TxFromAddress  string              `db:"tx_from_address"`
	TxToAddress    string              `db:"tx_to_address"`
	TxCreatedAt    pgtype.Timestamp    `db:"tx_created_at"`
	BlockRiskLevel string              `db:"bt_risk_level"`
	BlockScore     decimal.Decimal     `db:"bt_score"`
	BlockCreatedAt pgtype.Timestamp    `db:"bt_created_at"`

	// AML check that blocked the deposit
	AmlCheckID         uuid.UUID             `db:"aml_check_id"`
	AmlServiceSlug     models.AMLSlug        `db:"aml_service_slug"`
	AmlCheckExternalID string                `db:"aml_check_external_id"`
	AmlCheckStatus     models.AMLCheckStatus `db:"aml_check_status"`
	AmlCheckScore      decimal.Decimal       `db:"aml_check_score"`
	AmlCheckRiskLevel  *models.AmlRiskLevel  `db:"aml_check_risk_level"`
	AmlCheckCreatedAt  pgtype.Timestamp      `db:"aml_check_created_at"`
	AmlCheckUpdatedAt  pgtype.Timestamp      `db:"aml_check_updated_at"`

	// refund transfer, empty until the refund is approved
	TransferNumber    *int64                 `db:"transfer_number"`
	TransferStatus    *models.TransferStatus `db:"transfer_status"`
	TransferStage     *models.TransferStage  `db:"transfer_stage"`
	TransferStep      *string                `db:"transfer_step"`
	TransferTxHash    *string                `db:"transfer_tx_hash"`
	TransferAmount    decimal.NullDecimal    `db:"transfer_amount"`
	TransferAmountUsd decimal.NullDecimal    `db:"transfer_amount_usd"`
	TransferMessage   *string                `db:"transfer_message"`
	TransferCreatedAt pgtype.Timestamp       `db:"transfer_created_at"`
	TransferUpdatedAt pgtype.Timestamp       `db:"transfer_updated_at"`
}

var findAllowedOrderBy = map[string]string{
	"created_at": "rr.created_at",
	"updated_at": "rr.updated_at",
	"score":      "ac.score",
	"amount_usd": "t.amount_usd",
}

// newFindBuilder builds the joins and filters shared by the select and the count queries
func newFindBuilder(params FindParams, cols ...string) *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	sb.Select(cols...).
		From(sb.As("refund_requests", "rr")).
		Join(sb.As("stores", "s"), "s.id = rr.store_id").
		Join(sb.As("blocked_transactions", "bt"), "bt.id = rr.blocked_transaction_id").
		Join(sb.As("transactions", "t"), "t.id = bt.transaction_id").
		Join(sb.As("currencies", "c"), "c.id = t.currency_id").
		Join(sb.As("aml_checks", "ac"), "ac.id = bt.aml_check_id").
		Join(sb.As("aml_services", "asv"), "asv.id = ac.service_id").
		JoinWithOption(sqlbuilder.LeftJoin, sb.As("transfers", "tr"), "tr.id = rr.transfer_id").
		Where(sb.Equal("s.user_id", params.UserID))

	if len(params.Statuses) > 0 {
		sb.Where(sb.In("rr.status", toArgs(params.Statuses)...))
	}
	if params.Blockchain != nil {
		sb.Where(sb.Equal("t.blockchain", *params.Blockchain))
	}
	if params.TxHash != "" {
		sb.Where(sb.Or(sb.Equal("t.tx_hash", params.TxHash), sb.Equal("tr.tx_hash", params.TxHash)))
	}
	if params.FromAddress != "" {
		sb.Where(sb.Equal("t.from_address", params.FromAddress))
	}
	if params.DestinationAddress != "" {
		sb.Where(sb.Equal("rr.destination_address", params.DestinationAddress))
	}
	if params.DateFrom != nil {
		sb.Where(sb.GreaterEqualThan("rr.created_at", *params.DateFrom))
	}
	if params.DateTo != nil {
		sb.Where(sb.LessEqualThan("rr.created_at", *params.DateTo))
	}

	return sb
}

func (s *CustomQuerier) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithFullPagination[*FindRow], error) {
	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}

	orderBy := "rr.created_at"
	if params.OrderBy != "" {
		orderBy, err = storecmn.SafeOrderBy(params.OrderBy, findAllowedOrderBy)
		if err != nil {
			return nil, fmt.Errorf("invalid sort parameter: %w", err)
		}
	}

	sb := newFindBuilder(params,
		"rr.*",
		"t.id AS tx_id",
		"t.tx_hash AS tx_hash",
		"t.amount AS tx_amount",
		"t.amount_usd AS tx_amount_usd",
		"t.currency_id AS tx_currency_id",
		"c.code AS tx_currency_code",
		"t.blockchain AS tx_blockchain",
		"t.from_address AS tx_from_address",
		"t.to_address AS tx_to_address",
		"t.created_at AS tx_created_at",
		"bt.risk_level AS bt_risk_level",
		"bt.score AS bt_score",
		"bt.created_at AS bt_created_at",
		"ac.id AS aml_check_id",
		"asv.slug AS aml_service_slug",
		"ac.external_id AS aml_check_external_id",
		"ac.status AS aml_check_status",
		"ac.score AS aml_check_score",
		"ac.risk_level AS aml_check_risk_level",
		"ac.created_at AS aml_check_created_at",
		"ac.updated_at AS aml_check_updated_at",
		"tr.number AS transfer_number",
		"tr.status AS transfer_status",
		"tr.stage AS transfer_stage",
		"tr.step AS transfer_step",
		"tr.tx_hash AS transfer_tx_hash",
		"tr.amount AS transfer_amount",
		"tr.amount_usd AS transfer_amount_usd",
		"tr.message AS transfer_message",
		"tr.created_at AS transfer_created_at",
		"tr.updated_at AS transfer_updated_at",
	)

	direction := "DESC"
	if params.IsAscOrdering {
		direction = "ASC"
	}
	// id as a tie-breaker keeps pages stable when sort values are equal
	sb.OrderBy(orderBy+" "+direction+" NULLS LAST", "rr.id "+direction)
	sb.Limit(int(limit))   // #nosec
	sb.Offset(int(offset)) // #nosec

	items := make([]*FindRow, 0)
	query, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &items, query, args...); err != nil {
		return nil, fmt.Errorf("select refund requests: %w", err)
	}

	var total uint64
	countQuery, countArgs := newFindBuilder(params, "COUNT(rr.id)").Build()
	if err := pgxscan.Get(ctx, s.psql, &total, countQuery, countArgs...); err != nil {
		return nil, fmt.Errorf("count refund requests: %w", err)
	}

	var page uint64 = 1
	if params.Page != nil && *params.Page > 0 {
		page = uint64(*params.Page)
	}

	lastPage := uint64(math.Ceil(float64(total) / float64(limit)))
	if lastPage == 0 {
		lastPage = 1
	}

	return &storecmn.FindResponseWithFullPagination[*FindRow]{
		Items: items,
		Pagination: storecmn.FullPagingData{
			Total:    total,
			PageSize: uint64(limit),
			Page:     page,
			LastPage: lastPage,
		},
	}, nil
}

// toArgs converts a slice to sql args. sqlbuilder.Flatten can't be used: it splits uuid.UUID ([16]byte) into bytes
func toArgs[T any](items []T) []any {
	args := make([]any, len(items))
	for i, item := range items {
		args[i] = item
	}
	return args
}
