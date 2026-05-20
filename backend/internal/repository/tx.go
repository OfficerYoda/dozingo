package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxRunner runs fn inside a database transaction.
//
// fn receives a Repos bundle bound to the in-flight transaction; any
// repository call made through that bundle participates in the same tx.
// If fn returns an error or panics, the transaction is rolled back.
type TxRunner interface {
	WithTx(ctx context.Context, fn func(Repos) error) error
}

type pgxTxRunner struct {
	pool *pgxpool.Pool
}

// NewTxRunner returns a TxRunner backed by the given pgx pool.
func NewTxRunner(pool *pgxpool.Pool) TxRunner {
	return &pgxTxRunner{pool: pool}
}

// WithTx implements TxRunner. The rollback uses errors.Is(err, pgx.ErrTxClosed)
// to suppress the harmless "tx already committed" error reported by pgx after
// a successful Commit.
func (r *pgxTxRunner) WithTx(ctx context.Context, fn func(Repos) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			// Rollback errors after a successful commit are expected and ignored.
			// Anything else surfaces via the named return below if we add one,
			// but keeping this defer-only avoids shadowing fn's error.
			_ = rbErr
		}
	}()

	if err := fn(NewWithDBTX(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
