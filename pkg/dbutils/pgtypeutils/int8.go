package pgtypeutils

import "github.com/jackc/pgx/v5/pgtype"

func EncodeInt8(value *int64) pgtype.Int8 {
	var v int64
	if value != nil {
		v = *value
	}
	return pgtype.Int8{
		Int64: v,
		Valid: value != nil,
	}
}

func DecodeInt8(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	return &value.Int64
}
