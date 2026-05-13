package database

import (
	"context"

	"backend/internal/platform/config"
)

type DB interface {
	Ping(ctx context.Context) error
}

type DummyDB struct{}

func (d DummyDB) Ping(ctx context.Context) error {
	return nil
}

func RunMigrations(cfg *config.Config) error {
	return nil
}
