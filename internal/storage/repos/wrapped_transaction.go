package repos

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type DBTx interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

func BeginTxFunc(ctx context.Context, dbTx DBTx, txOptions pgx.TxOptions, fn func(pgx.Tx) error, opts ...Option) (err error) {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	if options.Tx != nil {
		return fn(options.Tx)
	}

	return pgx.BeginTxFunc(ctx, dbTx, txOptions, fn) //nolint:dbtxcheck
}
