package repo_notification_send_history

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/huandu/go-sqlbuilder"
)

func TestDeleteOldHistory(t *testing.T) {
	db := sqlbuilder.PostgreSQL.NewDeleteBuilder()

	db.DeleteFrom("notification_send_history").
		Where(db.GreaterThan("created_at", sqlbuilder.Raw("CURRENT_TIMESTAMP - INTERVAL '1 week'"))).Where("attempts >= 2")

	deleteSQL, args := db.Build()
	t.Log(deleteSQL)
	require.Equal(t, `DELETE FROM notification_send_history WHERE created_at > CURRENT_TIMESTAMP - INTERVAL '1 week' AND attempts >= 2`, deleteSQL)
	require.Empty(t, args)
}
