package database

import (
	"context"
	"database/sql"
)

type sqlTx interface {
	Commit() error
	Rollback() error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

type Beginner interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (sqlTx, error)
}

type SQLTxManager struct {
	db Beginner
}

func NewSQLTxManager(db Beginner) *SQLTxManager {
	return &SQLTxManager{db: db}
}

func (m *SQLTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if ExecutorFromContext(ctx, nil) != nil {
		return fn(ctx)
	}

	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	txCtx := context.WithValue(ctx, txKey{}, DBTX(tx))
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
