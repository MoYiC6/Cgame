package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

type testDBTX struct{}

func (testDBTX) Exec(query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (testDBTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (testDBTX) Query(query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (testDBTX) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (testDBTX) QueryRow(query string, args ...any) *sql.Row {
	return nil
}

func (testDBTX) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func TestExecutorFromContext_fallback(t *testing.T) {
	ctx := context.Background()
	fallback := testDBTX{}

	got := ExecutorFromContext(ctx, fallback)

	if got != fallback {
		t.Fatalf("expected fallback executor")
	}
}

func TestExecutorFromContext_stored(t *testing.T) {
	ctx := context.Background()
	stored := testDBTX{}
	fallback := testDBTX{}
	ctx = context.WithValue(ctx, txKey{}, stored)

	got := ExecutorFromContext(ctx, fallback)

	if got != stored {
		t.Fatalf("expected stored executor")
	}
}

func TestNoopTxManager_commit(t *testing.T) {
	mgr := NoopTxManager{}
	called := false

	err := mgr.WithinTx(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatalf("expected callback to be called")
	}
}

func TestNoopTxManager_rollback(t *testing.T) {
	mgr := NoopTxManager{}
	expectedErr := errors.New("rollback")

	err := mgr.WithinTx(context.Background(), func(ctx context.Context) error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected same error, got %v", err)
	}
}

func TestWithinTx_nested_reuse_outer(t *testing.T) {
	ctx := context.Background()
	outer := testDBTX{}
	ctx = context.WithValue(ctx, txKey{}, outer)
	mgr := NoopTxManager{}

	err := mgr.WithinTx(ctx, func(innerCtx context.Context) error {
		seen := ExecutorFromContext(innerCtx, testDBTX{})
		if seen != outer {
			t.Fatalf("expected nested WithinTx to reuse outer executor")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
