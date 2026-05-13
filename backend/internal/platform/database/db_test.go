package database

import (
	"context"
	"path/filepath"
	"testing"

	"backend/internal/platform/config"
)

func TestDummyDBPingReturnsNil(t *testing.T) {
	db := DummyDB{}
	if err := db.Ping(context.Background()); err != nil {
		t.Fatalf("Ping returned error: %v", err)
	}
}

func TestRunMigrationsAcceptsConfig(t *testing.T) {
	cfg, err := config.Load(filepath.Join("..", "..", "..", "configs", "config.test.yaml"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if err := RunMigrations(&cfg); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}
}
