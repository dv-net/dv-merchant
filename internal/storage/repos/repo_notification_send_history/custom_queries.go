package repo_notification_send_history

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
	"github.com/jackc/pgx/v5"
)

type ICustomQuerier interface {
	Querier
	DeleteOldHistory(context.Context) (int64, error)
	GetHistoryByUser(context.Context, uuid.UUID, GetHistoryByUserParams) (*storecmn.FindResponseWithFullPagination[*models.NotificationSendHistory], error)
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

type GetHistoryByUserParams struct {
	*storecmn.CommonFindParams
	IDs          []uuid.UUID
	Destinations []string
	Types        []models.NotificationType
	Channels     []models.DeliveryChannel
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
	SentFrom     *time.Time
	SentTo       *time.Time
	IsRoot       bool
}

func (s *CustomQuerier) GetHistoryByUser(ctx context.Context, userID uuid.UUID, params GetHistoryByUserParams) (*storecmn.FindResponseWithFullPagination[*models.NotificationSendHistory], error) {
	sb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	countSb := sqlbuilder.PostgreSQL.NewSelectBuilder()

	sb.Select("nsh.id", "nsh.destination", "nsh.message_text", "nsh.sender", "nsh.created_at", "nsh.updated_at", "nsh.sent_at", "nsh.type", "nsh.channel", "nsh.notification_send_queue_id").
		From("notification_send_history nsh").
		JoinWithOption(sqlbuilder.LeftJoin, "user_stores us", "nsh.store_id = us.store_id").
		JoinWithOption(sqlbuilder.LeftJoin, "notifications n", "n.type = nsh.type")

	countSb.Select("COUNT(*)").
		From("notification_send_history nsh").
		JoinWithOption(sqlbuilder.LeftJoin, "user_stores us", "nsh.store_id = us.store_id").
		JoinWithOption(sqlbuilder.LeftJoin, "notifications n", "n.type = nsh.type")

	// Add user filtering only if not root
	if !params.IsRoot {
		sb.Where(sb.Or(
			sb.Equal("us.user_id", userID),
			sb.Equal("nsh.user_id", userID),
		))
		countSb.Where(countSb.Or(
			countSb.Equal("us.user_id", userID),
			countSb.Equal("nsh.user_id", userID),
		))
	}

	// Filter out system notifications
	sb.Where(sb.IsNull("n.category"))
	countSb.Where(countSb.IsNull("n.category"))

	// Add type filtering
	if len(params.Types) > 0 {
		sb.Where(sb.In("nsh.type", sqlbuilder.Flatten(params.Types)...))
		countSb.Where(countSb.In("nsh.type", sqlbuilder.Flatten(params.Types)...))
	}

	// Add channel filtering
	if len(params.Channels) > 0 {
		sb.Where(sb.In("nsh.channel", sqlbuilder.Flatten(params.Channels)...))
		countSb.Where(countSb.In("nsh.channel", sqlbuilder.Flatten(params.Channels)...))
	}

	// Add ID filtering
	if len(params.IDs) > 0 {
		var idStrings []string
		for _, id := range params.IDs {
			idStrings = append(idStrings, id.String())
		}
		sb.Where(sb.In("nsh.id", sqlbuilder.Flatten(idStrings)...))
		countSb.Where(countSb.In("nsh.id", sqlbuilder.Flatten(idStrings)...))
	}

	// Add destination filtering
	if len(params.Destinations) > 0 {
		sb.Where(sb.In("nsh.destination", sqlbuilder.Flatten(params.Destinations)...))
		countSb.Where(countSb.In("nsh.destination", sqlbuilder.Flatten(params.Destinations)...))
	}

	// Add created_at date range filtering
	if params.CreatedFrom != nil {
		sb.Where(sb.GreaterEqualThan("nsh.created_at", *params.CreatedFrom))
		countSb.Where(countSb.GreaterEqualThan("nsh.created_at", *params.CreatedFrom))
	}
	if params.CreatedTo != nil {
		sb.Where(sb.LessEqualThan("nsh.created_at", *params.CreatedTo))
		countSb.Where(countSb.LessEqualThan("nsh.created_at", *params.CreatedTo))
	}

	// Add sent_at date range filtering
	if params.SentFrom != nil {
		sb.Where(sb.GreaterEqualThan("nsh.sent_at", *params.SentFrom))
		countSb.Where(countSb.GreaterEqualThan("nsh.sent_at", *params.SentFrom))
	}
	if params.SentTo != nil {
		sb.Where(sb.LessEqualThan("nsh.sent_at", *params.SentTo))
		countSb.Where(countSb.LessEqualThan("nsh.sent_at", *params.SentTo))
	}

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize, dbutils.WithMaxLimit(500))
	if err != nil {
		return nil, err
	}

	if params.OrderBy == "" {
		params.OrderBy = "nsh.created_at"
	} else {
		params.OrderBy = "nsh." + params.OrderBy
	}
	sb.OrderBy(params.OrderBy)
	if !params.IsAscOrdering {
		sb.Desc()
	}

	sb.Limit(int(limit))
	sb.Offset(int(offset))

	var items []*models.NotificationSendHistory
	sql, args := sb.Build()
	if err := pgxscan.Select(ctx, s.psql, &items, sql, args...); err != nil {
		return nil, fmt.Errorf("select: %w", err)
	}

	var totalCnt uint32
	pagingSQL, args := countSb.Build()
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

	return &storecmn.FindResponseWithFullPagination[*models.NotificationSendHistory]{
		Items: items,
		Pagination: storecmn.FullPagingData{
			Total:    uint64(totalCnt),
			PageSize: uint64(limit),
			Page:     uint64(currPage),
			LastPage: pagesCnt,
		},
	}, nil
}
