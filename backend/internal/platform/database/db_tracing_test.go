package database

import (
	"context"
	"testing"

	"backend/internal/platform/config"
)

func TestPgxPoolTracingHookRegistered(t *testing.T) {
	ctx := context.Background()
	pool, err := NewPgxPool(ctx, config.DBConfig{DSN: "postgres://user:pass@127.0.0.1:5432/app?sslmode=disable"})
	if err != nil {
		t.Fatalf("NewPgxPool returned error: %v", err)
	}
	defer pool.Close()

	if pool.Pool() == nil {
		t.Fatal("expected pool to be initialized")
	}
	if pool.Pool().Config().ConnConfig.Tracer == nil {
		t.Fatal("expected pgx pool tracer hook to be registered")
	}
}
