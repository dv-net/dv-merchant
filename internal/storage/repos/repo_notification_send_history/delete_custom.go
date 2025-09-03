package repo_notification_send_history

import (
	"context"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
)

func (s *CustomQuerier) DeleteOldHistory(ctx context.Context) (int64, error) {
	db := sqlbuilder.PostgreSQL.NewDeleteBuilder()

	db.DeleteFrom("notification_send_history").
		Where(db.LessThan("created_at", sqlbuilder.Raw("CURRENT_TIMESTAMP - INTERVAL '1 week'")))

	deleteSQL, args := db.Build()

	result, err := s.psql.Exec(ctx, deleteSQL, args...)
	if err != nil {
		return 0, fmt.Errorf("delete: %w", err)
	}

	return result.RowsAffected(), nil
}
