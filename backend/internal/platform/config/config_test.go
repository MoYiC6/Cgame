package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigLoadsAuthSection(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "..", "configs", "config.test.yaml"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Auth.Issuer != "backend" {
		t.Fatalf("expected auth issuer backend, got %q", cfg.Auth.Issuer)
	}
	if cfg.Auth.Cookie.Name != "refresh_token" {
		t.Fatalf("expected refresh cookie name refresh_token, got %q", cfg.Auth.Cookie.Name)
	}
	if cfg.Auth.JWT.Algorithm != "HS256" {
		t.Fatalf("expected jwt algorithm HS256, got %q", cfg.Auth.JWT.Algorithm)
	}
}

func TestLoadConfigRejectsProdWithoutJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("APP_CONFIG_PATH", filepath.Join("..", "..", "..", "configs", "config.prod.yaml"))
	t.Setenv("JWT_HMAC_SECRET", "")

	_, err := LoadConfig("")
	if err == nil {
		t.Fatal("expected error when prod jwt secret is missing")
	}

	if !strings.Contains(strings.ToLower(err.Error()), "jwt_hmac_secret") {
		t.Fatalf("expected jwt_hmac_secret validation error, got %v", err)
	}
}

func TestLoadFromFile(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "..", "configs", "config.test.yaml"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.App.Name != "backend-test" {
		t.Fatalf("expected app name backend-test, got %q", cfg.App.Name)
	}
	if cfg.Server.Addr != ":18080" {
		t.Fatalf("expected server addr :18080, got %q", cfg.Server.Addr)
	}
	if cfg.Log.Level != "debug" {
		t.Fatalf("expected log level debug, got %q", cfg.Log.Level)
	}
	if cfg.DB.Driver != "postgres" {
		t.Fatalf("expected DB driver postgres, got %q", cfg.DB.Driver)
	}
	if cfg.DB.MaxOpenConns != 16 {
		t.Fatalf("expected db max_open_conns 16, got %d", cfg.DB.MaxOpenConns)
	}
	if cfg.DB.MaxIdleConns != 8 {
		t.Fatalf("expected db max_idle_conns 8, got %d", cfg.DB.MaxIdleConns)
	}
	if cfg.DB.ConnMaxLifetimeSecs != 300 {
		t.Fatalf("expected db conn_max_lifetime_secs 300, got %d", cfg.DB.ConnMaxLifetimeSecs)
	}
	if cfg.Observability.TraceExporterType != "otlp" {
		t.Fatalf("expected trace_exporter_type otlp, got %q", cfg.Observability.TraceExporterType)
	}
	if cfg.Observability.ServiceName != "backend-test" {
		t.Fatalf("expected observability service_name backend-test, got %q", cfg.Observability.ServiceName)
	}
	if cfg.MQ.TopicPrefix != "backend.test" {
		t.Fatalf("expected topic prefix backend.test, got %q", cfg.MQ.TopicPrefix)
	}
	if cfg.Metrics.Enabled {
		t.Fatalf("expected metrics disabled by default in test config")
	}
}

func TestLoadRejectsMissingAppName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	content := []byte("app:\n  name: \"\"\nserver:\n  addr: \":8080\"\nlog:\n  level: info\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected Load to fail when app.name is empty")
	}
	if !strings.Contains(err.Error(), "app.name") {
		t.Fatalf("expected error to mention app.name, got %v", err)
	}
}

func TestLoadRejectsInvalidDBOrObservabilityConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.yaml")
	content := []byte(`app:
  name: backend-test
  env: test
server:
  addr: ":18080"
log:
  level: debug
db:
  driver: postgres
  dsn: "postgres://user:secret@localhost:5432/backend_test"
  max_open_conns: -1
  max_idle_conns: 5
  conn_max_lifetime_secs: 300
observability:
  trace_exporter_type: invalid-exporter
  trace_exporter_endpoint: "http://localhost:4318"
  service_name: backend-test
  service_version: "1.0.0"
  environment: test
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected Load to fail when db/observability config is invalid")
	}
	if !strings.Contains(err.Error(), "observability") && !strings.Contains(err.Error(), "db") {
		t.Fatalf("expected error to mention db or observability, got %v", err)
	}
}

func TestMaskedSummaryIncludesPhaseAFields(t *testing.T) {
	cfg := Config{
		App:             AppConfig{Name: "backend", Env: "test"},
		Server:          ServerConfig{Addr: ":18080"},
		Log:             LogConfig{Level: "debug"},
		DB:              DBConfig{Driver: "postgres", DSN: "postgres://user:secret@localhost:5432/backend_test"},
		Redis:           RedisConfig{Addr: "127.0.0.1:6379"},
		MQ:              MQConfig{Driver: "in-memory", TopicPrefix: "backend.test"},
		Metrics:         MetricsConfig{Enabled: true},
		CORS:            CORSConfig{AllowedOrigins: []string{"https://frontend.example.com"}},
		RateLimit:       RateLimitConfig{Requests: 100, WindowSecs: 60},
		SecurityHeaders: SecurityHeadersConfig{FrameOptions: "DENY"},
	}

	summary := cfg.MaskedSummary()
	if summary["metrics_enabled"] != "true" {
		t.Fatalf("expected metrics_enabled true, got %q", summary["metrics_enabled"])
	}
	if summary["cors_allowed_origins"] != "https://frontend.example.com" {
		t.Fatalf("expected cors_allowed_origins set, got %q", summary["cors_allowed_origins"])
	}
	if summary["rate_limit_requests"] != "100" {
		t.Fatalf("expected rate_limit_requests 100, got %q", summary["rate_limit_requests"])
	}
	if summary["rate_limit_window_secs"] != "60" {
		t.Fatalf("expected rate_limit_window_secs 60, got %q", summary["rate_limit_window_secs"])
	}
}

func TestLoadConfigAppliesEnvironmentOverrides(t *testing.T) {
	t.Setenv("APP_CONFIG_PATH", filepath.Join("..", "..", "..", "configs", "config.test.yaml"))
	t.Setenv("APP_NAME", "override-app")
	t.Setenv("APP_ENV", "dev")
	t.Setenv("SERVER_ADDR", ":19090")
	t.Setenv("LOG_LEVEL", "warn")
	t.Setenv("DB_DRIVER", "mysql")
	t.Setenv("DB_DSN", "mysql://u:p@tcp(localhost:3306)/backend")
	t.Setenv("DB_MAX_OPEN_CONNS", "32")
	t.Setenv("DB_MAX_IDLE_CONNS", "12")
	t.Setenv("DB_CONN_MAX_LIFETIME_SECS", "600")
	t.Setenv("OTEL_TRACE_EXPORTER_TYPE", "none")
	t.Setenv("OTEL_TRACE_EXPORTER_ENDPOINT", "http://collector:4318")
	t.Setenv("OTEL_SERVICE_NAME", "override-otel-service")
	t.Setenv("OTEL_SERVICE_VERSION", "2.0.0")
	t.Setenv("OTEL_ENVIRONMENT", "staging")
	t.Setenv("REDIS_ADDR", "redis:6379")

	cfg, err := LoadConfig("test")
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.App.Name != "override-app" {
		t.Fatalf("expected app name override-app, got %q", cfg.App.Name)
	}
	if cfg.App.Env != "dev" {
		t.Fatalf("expected env dev, got %q", cfg.App.Env)
	}
	if cfg.Server.Addr != ":19090" {
		t.Fatalf("expected server addr :19090, got %q", cfg.Server.Addr)
	}
	if cfg.Log.Level != "warn" {
		t.Fatalf("expected log level warn, got %q", cfg.Log.Level)
	}
	if cfg.DB.Driver != "mysql" {
		t.Fatalf("expected DB driver mysql, got %q", cfg.DB.Driver)
	}
	if cfg.DB.DSN != "mysql://u:p@tcp(localhost:3306)/backend" {
		t.Fatalf("expected DB DSN overridden, got %q", cfg.DB.DSN)
	}
	if cfg.DB.MaxOpenConns != 32 {
		t.Fatalf("expected DB max_open_conns 32, got %d", cfg.DB.MaxOpenConns)
	}
	if cfg.DB.MaxIdleConns != 12 {
		t.Fatalf("expected DB max_idle_conns 12, got %d", cfg.DB.MaxIdleConns)
	}
	if cfg.DB.ConnMaxLifetimeSecs != 600 {
		t.Fatalf("expected DB conn_max_lifetime_secs 600, got %d", cfg.DB.ConnMaxLifetimeSecs)
	}
	if cfg.Observability.TraceExporterType != "none" {
		t.Fatalf("expected trace_exporter_type none, got %q", cfg.Observability.TraceExporterType)
	}
	if cfg.Observability.TraceExporterEndpoint != "http://collector:4318" {
		t.Fatalf("expected trace_exporter_endpoint overridden, got %q", cfg.Observability.TraceExporterEndpoint)
	}
	if cfg.Observability.ServiceName != "override-otel-service" {
		t.Fatalf("expected service_name override-otel-service, got %q", cfg.Observability.ServiceName)
	}
	if cfg.Observability.ServiceVersion != "2.0.0" {
		t.Fatalf("expected service_version 2.0.0, got %q", cfg.Observability.ServiceVersion)
	}
	if cfg.Observability.Environment != "staging" {
		t.Fatalf("expected observability environment staging, got %q", cfg.Observability.Environment)
	}
	if cfg.Redis.Addr != "redis:6379" {
		t.Fatalf("expected redis addr overridden, got %q", cfg.Redis.Addr)
	}
}

func TestLoadConfigUsesAPPENVWhenArgumentEmpty(t *testing.T) {
	t.Setenv("APP_CONFIG_PATH", "")
	t.Setenv("APP_ENV", "test")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd returned error: %v", err)
	}
	backendRoot := filepath.Join(wd, "..", "..", "..")
	if err := os.Chdir(backendRoot); err != nil {
		t.Fatalf("Chdir returned error: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	})

	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.App.Env != "test" {
		t.Fatalf("expected app env test from APP_ENV, got %q", cfg.App.Env)
	}
	if cfg.Server.Addr != ":18080" {
		t.Fatalf("expected server addr from config.test.yaml, got %q", cfg.Server.Addr)
	}
}

func TestLocalConfigTargetsDockerDatastores(t *testing.T) {
	cfg, err := Load(filepath.Join("..", "..", "..", "configs", "config.local.yaml"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	const expectedDSN = "postgres://backend:backend@localhost:25432/Cgame?sslmode=disable"
	if cfg.DB.DSN != expectedDSN {
		t.Fatalf("expected local db dsn %q, got %q", expectedDSN, cfg.DB.DSN)
	}
	if cfg.Redis.Addr != "localhost:26379" {
		t.Fatalf("expected local redis addr localhost:26379, got %q", cfg.Redis.Addr)
	}
}
