package database

import (
	"context"
	"strings"
	"testing"

	"backend/internal/platform/config"
)

func TestDummyDBPingReturnsNil(t *testing.T) {
	db := DummyDB{}
	if err := db.Ping(context.Background()); err != nil {
		t.Fatalf("Ping returned error: %v", err)
	}
}

func TestNewPgxPoolRequiresDSN(t *testing.T) {
	_, err := NewPgxPool(context.Background(), config.DBConfig{})
	if err == nil {
		t.Fatal("expected missing dsn error")
	}
	if !strings.Contains(err.Error(), "dsn") {
		t.Fatalf("expected dsn error, got %v", err)
	}
}

func TestRunMigrationsRequiresConfig(t *testing.T) {
	err := RunMigrations(nil)
	if err == nil {
		t.Fatal("expected nil config error")
	}
	if !strings.Contains(err.Error(), "config") {
		t.Fatalf("expected config error, got %v", err)
	}
}

func TestRunMigrationsRequiresDSN(t *testing.T) {
	err := RunMigrations(&config.Config{})
	if err == nil {
		t.Fatal("expected missing dsn error")
	}
	if !strings.Contains(err.Error(), "dsn") {
		t.Fatalf("expected dsn error, got %v", err)
	}
}
