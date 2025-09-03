package pgtypeutils

import "github.com/jackc/pgx/v5/pgtype"

func EncodeBool(value *bool) pgtype.Bool {
	var v bool
	if value != nil {
		v = *value
	}
	return pgtype.Bool{
		Bool:  v,
		Valid: value != nil,
	}
}

func DecodeBool(value pgtype.Bool) *bool {
	if !value.Valid {
		return nil
	}
	return &value.Bool
}
