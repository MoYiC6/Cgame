package user

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRepositoryGetByEmailReturnsNormalizedUser(t *testing.T) {
	ctx := context.Background()
	h, err := newUserRepositoryHarness(ctx, t)
	if err != nil {
		t.Fatalf("newUserRepositoryHarness() error = %v", err)
	}
	defer func() {
		if err := h.Close(ctx); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	if err := h.ApplyMigrations(ctx); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	_, err = h.db.ExecContext(ctx, `
		INSERT INTO users (public_id, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, "usr_test_email", "admin@example.com", "hash", StatusActive, now, now)
	if err != nil {
		t.Fatalf("insert user error = %v", err)
	}

	repo := NewRepository(h.db)
	user, err := repo.GetByEmail(ctx, NormalizeEmail(" Admin@Example.com "))
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.PublicID != "usr_test_email" {
		t.Fatalf("expected public_id usr_test_email, got %q", user.PublicID)
	}
	if user.Email != "admin@example.com" {
		t.Fatalf("expected normalized email, got %q", user.Email)
	}
	if user.Status != StatusActive {
		t.Fatalf("expected status %q, got %q", StatusActive, user.Status)
	}
}

func TestRepositoryGetByIDReturnsUser(t *testing.T) {
	ctx := context.Background()
	h, err := newUserRepositoryHarness(ctx, t)
	if err != nil {
		t.Fatalf("newUserRepositoryHarness() error = %v", err)
	}
	defer func() {
		if err := h.Close(ctx); err != nil {
			t.Fatalf("Close() error = %v", err)
		}
	}()

	if err := h.ApplyMigrations(ctx); err != nil {
		t.Fatalf("ApplyMigrations() error = %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	var userID int64
	err = h.db.QueryRowContext(ctx, `
		INSERT INTO users (public_id, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, "usr_test_id", "member@example.com", "hash", StatusDisabled, now, now).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user returning id error = %v", err)
	}

	repo := NewRepository(h.db)
	user, err := repo.GetByID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.ID != userID {
		t.Fatalf("expected id %d, got %d", userID, user.ID)
	}
	if user.PublicID != "usr_test_id" {
		t.Fatalf("expected public_id usr_test_id, got %q", user.PublicID)
	}
	if user.Status != StatusDisabled {
		t.Fatalf("expected status %q, got %q", StatusDisabled, user.Status)
	}
}

type userRepositoryHarness struct {
	db        *sql.DB
	container *postgres.PostgresContainer
}

func newUserRepositoryHarness(ctx context.Context, t *testing.T) (*userRepositoryHarness, error) {
	t.Helper()
	configureUserRepositoryDockerEnv(t)

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("backend_test"),
		postgres.WithUsername("backend"),
		postgres.WithPassword("backend"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
	)
	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	connString, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("postgres connection string: %w", err)
	}

	db, err := sql.Open("pgx", connString)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("sql open: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &userRepositoryHarness{db: db, container: container}, nil
}

func (h *userRepositoryHarness) ApplyMigrations(ctx context.Context) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	migrationsDir, err := migrationsDir()
	if err != nil {
		return err
	}
	return goose.UpContext(ctx, h.db, migrationsDir)
}

func (h *userRepositoryHarness) Close(ctx context.Context) error {
	var errs []string
	if h.db != nil {
		if err := h.db.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if h.container != nil {
		if err := h.container.Terminate(ctx); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("close harness: %s", strings.Join(errs, "; "))
	}
	return nil
}

func configureUserRepositoryDockerEnv(t *testing.T) {
	t.Helper()
	if endpoint := strings.TrimSpace(os.Getenv("DOCKER_HOST")); endpoint != "" {
		t.Setenv("DOCKER_HOST", endpoint)
	}
	if socketOverride := strings.TrimSpace(os.Getenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE")); socketOverride != "" {
		t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", socketOverride)
	}
	colimaSocket := filepath.Join(os.Getenv("HOME"), ".colima", "default", "docker.sock")
	if _, err := os.Stat(colimaSocket); err == nil {
		t.Setenv("DOCKER_HOST", "unix://"+colimaSocket)
		t.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/var/run/docker.sock")
	}
}

func migrationsDir() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve runtime caller for migrations dir")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")), nil
}
