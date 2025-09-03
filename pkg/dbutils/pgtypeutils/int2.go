package pgtypeutils

import "github.com/jackc/pgx/v5/pgtype"

func EncodeInt2(value *int16) pgtype.Int2 {
	var v int16
	if value != nil {
		v = *value
	}
	return pgtype.Int2{
		Int16: v,
		Valid: value != nil,
	}
}

func DecodeInt2(value pgtype.Int2) *int16 {
	if !value.Valid {
		return nil
	}
	return &value.Int16
}
