package database

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"backend/internal/platform/config"
	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const defaultMigrationsDir = "migrations"

type DB interface {
	Ping(ctx context.Context) error
}

type PgxPool struct {
	pool *pgxpool.Pool
}

func NewPgxPool(ctx context.Context, cfg config.DBConfig) (*PgxPool, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if cfg.DSN == "" {
		return nil, fmt.Errorf("db dsn is required")
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse pgx pool config: %w", err)
	}
	if cfg.MaxOpenConns > 0 {
		poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	}
	if cfg.ConnMaxLifetimeSecs > 0 {
		poolConfig.MaxConnLifetime = time.Duration(cfg.ConnMaxLifetimeSecs) * time.Second
	}
	poolConfig.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	return &PgxPool{pool: pool}, nil
}

func (p *PgxPool) Ping(ctx context.Context) error {
	if p == nil || p.pool == nil {
		return fmt.Errorf("pgx pool is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return p.pool.Ping(ctx)
}

func (p *PgxPool) Close() {
	if p == nil || p.pool == nil {
		return
	}
	p.pool.Close()
}

func (p *PgxPool) Shutdown(ctx context.Context) error {
	p.Close()
	return nil
}

func (p *PgxPool) Pool() *pgxpool.Pool {
	if p == nil {
		return nil
	}
	return p.pool
}

type DummyDB struct{}

func (d DummyDB) Ping(ctx context.Context) error {
	return nil
}

func RunMigrations(cfg *config.Config) error {
	return RunMigrationsFrom(cfg, defaultMigrationsDir)
}

func RunMigrationsFrom(cfg *config.Config, dir string) error {
	if cfg == nil {
		return fmt.Errorf("config is required")
	}
	if cfg.DB.DSN == "" {
		return fmt.Errorf("db dsn is required")
	}
	if dir == "" {
		dir = defaultMigrationsDir
	}

	db, err := sql.Open("pgx", cfg.DB.DSN)
	if err != nil {
		return fmt.Errorf("open migration db: %w", err)
	}
	defer db.Close()

	if cfg.DB.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	}
	if cfg.DB.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	}
	if cfg.DB.ConnMaxLifetimeSecs > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetimeSecs) * time.Second)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping migration db: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.UpContext(context.Background(), db, filepath.Clean(dir)); err != nil {
		return fmt.Errorf("apply migrations from %s: %w", filepath.Clean(dir), err)
	}
	return nil
}
