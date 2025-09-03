package pgtypeutils

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ParseUUID(id pgtype.UUID) (uuid.UUID, error) {
	return uuid.ParseBytes(id.Bytes[:])
}

func EncodeUUID(id uuid.UUID) (pgtype.UUID, error) {
	uid := pgtype.UUID{}
	err := uid.Scan(id.String())
	return uid, err
}

func IsUUIDNil(id pgtype.UUID) (bool, error) {
	if !id.Valid {
		return false, nil
	}

	uid, err := uuid.ParseBytes(id.Bytes[:])
	if err != nil {
		return false, err
	}

	return uid == uuid.Nil, nil
}
