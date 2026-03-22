package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

type Tx struct {
	pool *pgxpool.Pool
}

func NewTx(pool *pgxpool.Pool) *Tx {
	return &Tx{pool: pool}
}

func (t *Tx) Transactional(ctx context.Context, f func(ctx context.Context) error) error {
	if tx := t.getTx(ctx); tx != nil {
		return f(ctx)
	}

	tx, err := t.pool.Begin(ctx)
	if err != nil {
		return err
	}

	ctx = t.setTx(ctx, tx)

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := f(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (t *Tx) setTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// getTx извлекает транзакцию из контекста.
func (t *Tx) getTx(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(txKey{}).(pgx.Tx)
	return tx
}
