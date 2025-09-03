package pgtypeutils

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func EncodeTime(value time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:  value,
		Valid: value != time.Time{},
	}
}

func DecodeTime(value pgtype.Timestamp) *time.Time {
	var t *time.Time
	if value.Valid {
		t = &value.Time
	}
	return t
}
