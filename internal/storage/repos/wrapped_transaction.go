package repos

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	txRetryMaxAttempts = 5
	txRetryBaseDelay   = 10 * time.Millisecond
)

type DBTx interface {
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
}

// BeginTxFunc runs fn in a transaction. For SERIALIZABLE and REPEATABLE READ
// transactions it retries fn on serialization failures and deadlocks, as Postgres
// requires. fn must therefore have no side effects outside the database.
func BeginTxFunc(ctx context.Context, dbTx DBTx, txOptions pgx.TxOptions, fn func(pgx.Tx) error, opts ...Option) (err error) {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}

	if options.Tx != nil {
		// Nested call: the outer transaction owns retries.
		return fn(options.Tx)
	}

	if !isRetryableIsoLevel(txOptions.IsoLevel) {
		return pgx.BeginTxFunc(ctx, dbTx, txOptions, fn) //nolint:dbtxcheck
	}

	for attempt := 1; ; attempt++ {
		err = pgx.BeginTxFunc(ctx, dbTx, txOptions, fn) //nolint:dbtxcheck
		if err == nil || attempt >= txRetryMaxAttempts || !isRetryableTxError(err) {
			return err
		}

		// Exponential backoff with jitter: 5-10ms, 10-20ms, 20-40ms, 40-80ms.
		delay := txRetryBaseDelay << (attempt - 1)
		delay = delay/2 + rand.N(delay/2+1)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return err
		case <-timer.C:
		}
	}
}

func isRetryableIsoLevel(level pgx.TxIsoLevel) bool {
	return level == pgx.Serializable || level == pgx.RepeatableRead
}

func isRetryableTxError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	switch pgErr.Code {
	case "40001", // serialization_failure
		"40P01": // deadlock_detected
		return true
	}
	return false
}
