package pgerror

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func ParseError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return &UniqueConstraintError{
				Constraint: pgErr.ConstraintName,
				Table:      pgErr.TableName,
				Detail:     pgErr.Detail,
			}
		case "02000":
			return &NotFoundError{
				Table:  pgErr.TableName,
				Detail: pgErr.Detail,
			}
		}
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return &NotFoundError{
			Detail: "no rows found in query",
		}
	}
	return err
}

type UniqueConstraintError struct {
	Constraint string
	Table      string
	Detail     string
}

func (e *UniqueConstraintError) Error() string {
	return "unique constraint violation: " + e.Detail
}

type NotFoundError struct {
	Table  string
	Detail string
}

func (e *NotFoundError) Error() string {
	return "record not found: " + e.Detail
}
