package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	if cfg.MQ.TopicPrefix != "backend.test" {
		t.Fatalf("expected topic prefix backend.test, got %q", cfg.MQ.TopicPrefix)
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

func TestMaskedSummaryHidesSecrets(t *testing.T) {
	cfg := Config{
		App:    AppConfig{Name: "backend", Env: "test"},
		Server: ServerConfig{Addr: ":18080"},
		Log:    LogConfig{Level: "debug"},
		DB:     DBConfig{Driver: "postgres", DSN: "postgres://user:secret@localhost:5432/backend_test"},
		Redis:  RedisConfig{Addr: "127.0.0.1:6379"},
		MQ:     MQConfig{Driver: "in-memory", TopicPrefix: "backend.test"},
	}

	summary := cfg.MaskedSummary()
	if strings.Contains(summary["db_dsn"], "secret") {
		t.Fatalf("expected masked db_dsn, got %q", summary["db_dsn"])
	}
	if summary["app_name"] != "backend" {
		t.Fatalf("expected app_name backend, got %q", summary["app_name"])
	}
}

func TestLoadConfigAppliesEnvironmentOverrides(t *testing.T) {
	t.Setenv("APP_CONFIG_PATH", filepath.Join("..", "..", "..", "configs", "config.test.yaml"))
	t.Setenv("APP_NAME", "override-app")
	t.Setenv("APP_ENV", "dev")
	t.Setenv("SERVER_ADDR", ":19090")
	t.Setenv("LOG_LEVEL", "warn")

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
}
