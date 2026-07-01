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
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const defaultMigrationsDir = "migrations"

type DB interface {
	Ping(ctx context.Context) error
}

type PgxPool struct {
	pool *pgxpool.Pool
}

type SQLDB struct {
	db *sql.DB
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

func (p *PgxPool) SQLDB() (*SQLDB, error) {
	if p == nil || p.pool == nil {
		return nil, fmt.Errorf("pgx pool is nil")
	}
	db := stdlib.OpenDBFromPool(p.pool)
	return &SQLDB{db: db}, nil
}

func NewSQLDB(cfg config.DBConfig) (*SQLDB, error) {
	if cfg.DSN == "" {
		return nil, fmt.Errorf("db dsn is required")
	}
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open sql db: %w", err)
	}
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetimeSecs > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeSecs) * time.Second)
	}
	return &SQLDB{db: db}, nil
}

func (d *SQLDB) Ping(ctx context.Context) error {
	if d == nil || d.db == nil {
		return fmt.Errorf("sql db is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return d.db.PingContext(ctx)
}

func (d *SQLDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (sqlTx, error) {
	if d == nil || d.db == nil {
		return nil, fmt.Errorf("sql db is nil")
	}
	return d.db.BeginTx(ctx, opts)
}

func (d *SQLDB) Exec(query string, args ...any) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *SQLDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

func (d *SQLDB) Query(query string, args ...any) (*sql.Rows, error) {
	return d.db.Query(query, args...)
}

func (d *SQLDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d *SQLDB) QueryRow(query string, args ...any) *sql.Row {
	return d.db.QueryRow(query, args...)
}

func (d *SQLDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d *SQLDB) Close() error {
	if d == nil || d.db == nil {
		return nil
	}
	return d.db.Close()
}

func (d *SQLDB) Shutdown(ctx context.Context) error {
	_ = ctx
	return d.Close()
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
	defer func() {
		_ = db.Close()
	}()

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
