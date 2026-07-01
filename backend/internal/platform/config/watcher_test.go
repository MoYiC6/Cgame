package config_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"backend/internal/platform/config"
)

func TestWatcherReloadsOnFileChange(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	t.Setenv("JWT_HMAC_SECRET", "12345678901234567890123456789012")

	initialConfig := `
app:
  name: test-app
  env: local
server:
  addr: ":8080"
log:
  level: info
db:
  driver: pgx
  dsn: postgres://localhost/test
observability:
  service_name: test-service
auth:
  issuer: test-issuer
  audience: test-audience
  access_token_ttl: 15m
  refresh_token_ttl: 168h
  login:
    max_failed_attempts: 5
    failed_window: 5m
    lock_duration: 15m
  cookie:
    name: test-cookie
    path: /
    same_site: lax
  jwt:
    algorithm: HS256
    key_id: test-key
`
	if err := os.WriteFile(cfgPath, []byte(initialConfig), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	changeCount := 0
	onChange := func(cfg *config.Config) error {
		changeCount++
		return nil
	}

	watcher, err := config.NewWatcher(cfgPath, onChange, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("NewWatcher() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = watcher.Start(ctx)
	}()

	// wait for watcher to be ready
	<-time.After(100 * time.Millisecond)

	newConfig := strings.Replace(initialConfig, "level: info", "level: debug", 1)
	if err := os.WriteFile(cfgPath, []byte(newConfig), 0644); err != nil {
		t.Fatalf("write updated config: %v", err)
	}

	select {
	case <-time.After(2 * time.Second):
		if changeCount != 1 {
			t.Fatalf("onChange called %d times, want 1", changeCount)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for config reload")
	}
}
