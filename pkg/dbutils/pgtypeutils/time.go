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

func DecodeTime(value pgtype.Timestamp) time.Time {
	if !value.Valid {
		return time.Time{}
	}
	return value.Time
}
