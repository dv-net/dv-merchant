package models

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserInfo struct {
	ID                uuid.UUID        `db:"id" json:"id"`
	Email             string           `db:"email" json:"email" validate:"required,email"`
	EmailVerifiedAt   pgtype.Timestamp `db:"email_verified_at" json:"email_verified_at"`
	Location          string           `db:"location" json:"location" validate:"required,timezone"`
	Language          string           `db:"language" json:"language"`
	RateSource        string           `db:"rate_source" json:"rate_source"`
	ProcessingOwnerID uuid.NullUUID    `db:"processing_owner_id" json:"processing_owner_id"`
	CreatedAt         pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt         pgtype.Timestamp `db:"updated_at" json:"updated_at"`
}
