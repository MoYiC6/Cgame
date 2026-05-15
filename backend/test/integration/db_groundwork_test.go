package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"backend/internal/bootstrap"
	"backend/internal/platform/config"
	"backend/internal/platform/database"
	dbgen "backend/internal/platform/database/generated"
	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestDatabaseGroundwork(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx := context.Background()
	harness, err := newDatabaseGroundworkHarness(ctx, t)
	if err != nil {
		t.Fatalf("newDatabaseGroundworkHarness returned error: %v", err)
	}
	defer func() {
		if err := harness.Close(ctx); err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	}()

	if err := harness.ApplyMigrations(ctx); err != nil {
		t.Fatalf("ApplyMigrations returned error: %v", err)
	}

	queries := dbgen.New(harness.db.Pool())

	runID := fmt.Sprintf("run-%d", time.Now().UnixNano())
	if err := queries.InsertRuntimeProbe(ctx, dbgen.InsertRuntimeProbeParams{RunID: runID, ProbeName: "db-groundwork"}); err != nil {
		t.Fatalf("InsertRuntimeProbe returned error: %v", err)
	}

	count, err := queries.CountRuntimeProbesByRunID(ctx, runID)
	if err != nil {
		t.Fatalf("CountRuntimeProbesByRunID returned error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected probe count 1, got %d", count)
	}

	if err := assertReadinessUsesRealDBPing(harness.db); err != nil {
		t.Fatalf("assertReadinessUsesRealDBPing returned error: %v", err)
	}

	if err := queries.DeleteRuntimeProbesByRunID(ctx, runID); err != nil {
		t.Fatalf("DeleteRuntimeProbesByRunID returned error: %v", err)
	}

	count, err = queries.CountRuntimeProbesByRunID(ctx, runID)
	if err != nil {
		t.Fatalf("CountRuntimeProbesByRunID after delete returned error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected probe count 0 after delete, got %d", count)
	}
}

func TestDatabaseGroundworkFailsOnBrokenMigration(t *testing.T) {
	ctx := context.Background()
	harness, err := newDatabaseGroundworkHarness(ctx, t)
	if err != nil {
		t.Fatalf("newDatabaseGroundworkHarness returned error: %v", err)
	}
	defer func() {
		if err := harness.Close(ctx); err != nil {
			t.Fatalf("Close returned error: %v", err)
		}
	}()

	err = harness.ApplyMigrationsFrom(ctx, filepath.Join("testdata", "db_groundwork", "migrations", "broken"))
	if err == nil {
		t.Fatal("expected broken migrations to fail")
	}
	lower := strings.ToLower(err.Error())
	if !strings.Contains(lower, "syntax") && !strings.Contains(lower, "creat") {
		t.Fatalf("expected syntax-related migration error, got %v", err)
	}
}

type databaseGroundworkHarness struct {
	sqlDB     *sql.DB
	db        *database.PgxPool
	container *postgres.PostgresContainer
}

func newDatabaseGroundworkHarness(ctx context.Context, t *testing.T) (*databaseGroundworkHarness, error) {
	t.Helper()
	configureTestcontainersDockerEnvironment(t)

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
		if terminateErr := container.Terminate(ctx); terminateErr != nil {
			return nil, fmt.Errorf("connection string error: %w; terminate error: %v", err, terminateErr)
		}
		return nil, fmt.Errorf("postgres connection string: %w", err)
	}

	sqlDB, err := sql.Open("pgx", connString)
	if err != nil {
		if terminateErr := container.Terminate(ctx); terminateErr != nil {
			return nil, fmt.Errorf("sql open error: %w; terminate error: %v", err, terminateErr)
		}
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(pingCtx); err != nil {
		_ = sqlDB.Close()
		if terminateErr := container.Terminate(ctx); terminateErr != nil {
			return nil, fmt.Errorf("ping error: %w; terminate error: %v", err, terminateErr)
		}
		return nil, fmt.Errorf("ping postgres connection: %w", err)
	}

	pool, err := database.NewPgxPool(ctx, config.DBConfig{DSN: connString, MaxOpenConns: 16, ConnMaxLifetimeSecs: 300})
	if err != nil {
		_ = sqlDB.Close()
		if terminateErr := container.Terminate(ctx); terminateErr != nil {
			return nil, fmt.Errorf("new pgx pool error: %w; terminate error: %v", err, terminateErr)
		}
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		_ = sqlDB.Close()
		if terminateErr := container.Terminate(ctx); terminateErr != nil {
			return nil, fmt.Errorf("pgx pool ping error: %w; terminate error: %v", err, terminateErr)
		}
		return nil, fmt.Errorf("ping pgx pool: %w", err)
	}

	return &databaseGroundworkHarness{sqlDB: sqlDB, db: pool, container: container}, nil
}

func (h *databaseGroundworkHarness) ApplyMigrations(ctx context.Context) error {
	return h.ApplyMigrationsFrom(ctx, filepath.Join("testdata", "db_groundwork", "migrations", "happy"))
}

func (h *databaseGroundworkHarness) ApplyMigrationsFrom(ctx context.Context, dir string) error {
	if h == nil || h.sqlDB == nil {
		return fmt.Errorf("database harness is not initialized")
	}
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set goose dialect: %w", err)
	}
	if err := goose.UpContext(ctx, h.sqlDB, dir); err != nil {
		return fmt.Errorf("apply migrations from %s: %w", dir, err)
	}
	return nil
}

func (h *databaseGroundworkHarness) Close(ctx context.Context) error {
	if h == nil {
		return nil
	}
	var errs []string
	if h.db != nil {
		h.db.Close()
	}
	if h.sqlDB != nil {
		if err := h.sqlDB.Close(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if h.container != nil {
		if err := h.container.Terminate(ctx); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("close database groundwork harness: %s", strings.Join(errs, "; "))
	}
	return nil
}

func assertReadinessUsesRealDBPing(db database.DB) error {
	recording := &recordingDB{db: db}
	deps := bootstrap.Dependencies{
		Config: config.Config{
			App:    config.AppConfig{Name: "backend-test", Env: "test"},
			Server: config.ServerConfig{Addr: ":18080"},
			Log:    config.LogConfig{Level: "debug"},
		},
		Logger:     logger.New("debug", io.Discard),
		Tracer:     observability.NewNoopTracer(),
		Propagator: observability.NewNoopPropagator(),
		DB:         recording,
	}

	engine := bootstrap.NewAPIEngine(deps)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	engine.ServeHTTP(recorder, request)

	if recording.pingCalls != 1 {
		return fmt.Errorf("expected readiness to call DB ping once, got %d", recording.pingCalls)
	}
	if recorder.Code != http.StatusOK {
		return fmt.Errorf("expected /readyz 200, got %d", recorder.Code)
	}

	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		return fmt.Errorf("decode readiness response: %w", err)
	}

	data, ok := body.Data.(map[string]any)
	if !ok {
		return fmt.Errorf("expected readiness data map, got %#v", body.Data)
	}
	if data["status"] != "ok" {
		return fmt.Errorf("expected readiness status ok, got %#v", data["status"])
	}
	return nil
}

type recordingDB struct {
	db        database.DB
	pingCalls int
}

func (d *recordingDB) Ping(ctx context.Context) error {
	if d == nil || d.db == nil {
		return fmt.Errorf("db is nil")
	}
	d.pingCalls++
	return d.db.Ping(ctx)
}

func configureTestcontainersDockerEnvironment(t *testing.T) {
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
