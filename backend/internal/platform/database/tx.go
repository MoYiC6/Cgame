package database

import (
	"context"
	"database/sql"
	"time"
)

type TxOption struct {
	ReadOnly bool
	Label    string
	Timeout  time.Duration
}

type DBTX interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error, opts ...TxOption) error
}

type txKey struct{}

func ExecutorFromContext(ctx context.Context, fallback DBTX) DBTX {
	if ctx == nil {
		return fallback
	}

	exec, ok := ctx.Value(txKey{}).(DBTX)
	if !ok || exec == nil {
		return fallback
	}

	return exec
}

type NoopTxManager struct{}

func (NoopTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error, opts ...TxOption) error {
	return fn(ctx)
}
