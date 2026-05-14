package database

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

type fakeBeginner struct {
	tx         sqlTx
	beginErr   error
	beginCalls int
}

func (f *fakeBeginner) BeginTx(ctx context.Context, opts *sql.TxOptions) (sqlTx, error) {
	f.beginCalls++
	if f.beginErr != nil {
		return nil, f.beginErr
	}
	return f.tx, nil
}

type fakeTx struct {
	commitCalls   int
	rollbackCalls int
}

func (f *fakeTx) Commit() error {
	f.commitCalls++
	return nil
}

func (f *fakeTx) Rollback() error {
	f.rollbackCalls++
	return nil
}

func (f *fakeTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (f *fakeTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (f *fakeTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}

func (f *fakeTx) Exec(query string, args ...any) (sql.Result, error) {
	return nil, nil
}

func (f *fakeTx) Query(query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}

func (f *fakeTx) QueryRow(query string, args ...any) *sql.Row {
	return nil
}

func TestTxManagerCommitsOnSuccess(t *testing.T) {
	tx := &fakeTx{}
	beginner := &fakeBeginner{tx: tx}
	mgr := NewSQLTxManager(beginner)

	err := mgr.WithinTx(context.Background(), func(ctx context.Context) error {
		if ExecutorFromContext(ctx, nil) != tx {
			t.Fatalf("expected callback context to contain transaction executor")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if beginner.beginCalls != 1 {
		t.Fatalf("expected one BeginTx call, got %d", beginner.beginCalls)
	}
	if tx.commitCalls != 1 {
		t.Fatalf("expected one Commit call, got %d", tx.commitCalls)
	}
	if tx.rollbackCalls != 0 {
		t.Fatalf("expected no Rollback calls, got %d", tx.rollbackCalls)
	}
}

func TestTxManagerRollsBackOnError(t *testing.T) {
	tx := &fakeTx{}
	beginner := &fakeBeginner{tx: tx}
	mgr := NewSQLTxManager(beginner)
	expectedErr := errors.New("callback failed")

	err := mgr.WithinTx(context.Background(), func(ctx context.Context) error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected callback error, got %v", err)
	}
	if beginner.beginCalls != 1 {
		t.Fatalf("expected one BeginTx call, got %d", beginner.beginCalls)
	}
	if tx.rollbackCalls != 1 {
		t.Fatalf("expected one Rollback call, got %d", tx.rollbackCalls)
	}
	if tx.commitCalls != 0 {
		t.Fatalf("expected no Commit calls, got %d", tx.commitCalls)
	}
}

func TestTxManagerRollsBackOnPanic(t *testing.T) {
	tx := &fakeTx{}
	beginner := &fakeBeginner{tx: tx}
	mgr := NewSQLTxManager(beginner)
	panicValue := "boom"

	defer func() {
		r := recover()
		if r != panicValue {
			t.Fatalf("expected re-panic value %q, got %v", panicValue, r)
		}
		if beginner.beginCalls != 1 {
			t.Fatalf("expected one BeginTx call, got %d", beginner.beginCalls)
		}
		if tx.rollbackCalls != 1 {
			t.Fatalf("expected one Rollback call, got %d", tx.rollbackCalls)
		}
		if tx.commitCalls != 0 {
			t.Fatalf("expected no Commit calls, got %d", tx.commitCalls)
		}
	}()

	_ = mgr.WithinTx(context.Background(), func(ctx context.Context) error {
		panic(panicValue)
	})
}

func TestTxManagerHandlesContextCancel(t *testing.T) {
	beginErr := context.Canceled
	beginner := &fakeBeginner{beginErr: beginErr}
	mgr := NewSQLTxManager(beginner)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := mgr.WithinTx(ctx, func(ctx context.Context) error {
		t.Fatalf("callback must not be called when BeginTx fails")
		return nil
	})

	if !errors.Is(err, beginErr) {
		t.Fatalf("expected begin error, got %v", err)
	}
	if beginner.beginCalls != 1 {
		t.Fatalf("expected one BeginTx call, got %d", beginner.beginCalls)
	}
}

func TestTxManagerNestedReusesOuterTransaction(t *testing.T) {
	outerTx := &fakeTx{}
	beginner := &fakeBeginner{tx: outerTx}
	mgr := NewSQLTxManager(beginner)

	err := mgr.WithinTx(context.Background(), func(outerCtx context.Context) error {
		outerExecutor := ExecutorFromContext(outerCtx, nil)
		if outerExecutor != outerTx {
			t.Fatalf("expected outer transaction executor")
		}

		return mgr.WithinTx(outerCtx, func(innerCtx context.Context) error {
			innerExecutor := ExecutorFromContext(innerCtx, nil)
			if innerExecutor != outerExecutor {
				t.Fatalf("expected inner transaction to reuse outer executor")
			}
			return nil
		})
	})

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if beginner.beginCalls != 1 {
		t.Fatalf("expected nested transaction to begin once, got %d", beginner.beginCalls)
	}
	if outerTx.commitCalls != 1 {
		t.Fatalf("expected outer transaction to commit once, got %d", outerTx.commitCalls)
	}
	if outerTx.rollbackCalls != 0 {
		t.Fatalf("expected no rollback calls, got %d", outerTx.rollbackCalls)
	}
}
