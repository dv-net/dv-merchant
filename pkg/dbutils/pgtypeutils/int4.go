package pgtypeutils

import "github.com/jackc/pgx/v5/pgtype"

func EncodeInt4(value *int32) pgtype.Int4 {
	var v int32
	if value != nil {
		v = *value
	}
	return pgtype.Int4{
		Int32: v,
		Valid: value != nil,
	}
}

func DecodeInt4(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	return &value.Int32
}
