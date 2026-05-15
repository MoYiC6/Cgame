package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/config"
	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type stubDB struct {
	pingErr error
}

var _ interface{ Ping(context.Context) error } = stubDB{}

func (s stubDB) Ping(ctx context.Context) error {
	return s.pingErr
}

func TestNewAPIEngineRegistersHealthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := Dependencies{
		Config: config.Config{
			App:    config.AppConfig{Name: "backend-test", Env: "test"},
			Server: config.ServerConfig{Addr: ":18080"},
			Log:    config.LogConfig{Level: "debug"},
		},
		Logger:     logger.New("debug", io.Discard),
		Tracer:     observability.NewNoopTracer(),
		Propagator: observability.NewNoopPropagator(),
		DB:         stubDB{},
	}

	engine := NewAPIEngine(deps)

	for _, path := range []string{"/healthz", "/readyz"} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)
		engine.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", path, recorder.Code)
		}

		var body response.APIResponse
		if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal returned error: %v", err)
		}
		if body.Code != "OK" {
			t.Fatalf("expected code OK for %s, got %q", path, body.Code)
		}
		if body.RequestID == "" {
			t.Fatalf("expected request_id for %s", path)
		}
		if body.TraceID == "" {
			t.Fatalf("expected trace_id for %s", path)
		}
	}
}

func TestReadyzReturnsOKWhenDBHealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := Dependencies{
		Config: config.Config{
			App:    config.AppConfig{Name: "backend-test", Env: "test"},
			Server: config.ServerConfig{Addr: ":18080"},
			Log:    config.LogConfig{Level: "debug"},
		},
		Logger:     logger.New("debug", io.Discard),
		Tracer:     observability.NewNoopTracer(),
		Propagator: observability.NewNoopPropagator(),
		DB:         stubDB{},
	}

	engine := NewAPIEngine(deps)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	request.Header.Set("X-Trace-ID", "trace-ready-healthy")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != "OK" {
		t.Fatalf("expected code OK, got %q", body.Code)
	}
	if body.RequestID == "" {
		t.Fatalf("expected request_id")
	}
	if body.TraceID != "trace-ready-healthy" {
		t.Fatalf("expected trace_id trace-ready-healthy, got %q", body.TraceID)
	}
	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got %#v", body.Data)
	}
	if data["status"] != "ok" {
		t.Fatalf("expected status ok, got %#v", data["status"])
	}
	if _, found := data["dependencies"]; found {
		t.Fatalf("did not expect dependencies field, got %#v", data)
	}
}

func TestReadyzReturnsServiceUnavailableWhenDBFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := Dependencies{
		Config: config.Config{
			App:    config.AppConfig{Name: "backend-test", Env: "test"},
			Server: config.ServerConfig{Addr: ":18080"},
			Log:    config.LogConfig{Level: "debug"},
		},
		Logger:     logger.New("debug", io.Discard),
		Tracer:     observability.NewNoopTracer(),
		Propagator: observability.NewNoopPropagator(),
		DB:         stubDB{pingErr: errors.New("db down")},
	}

	engine := NewAPIEngine(deps)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	request.Header.Set("X-Trace-ID", "trace-ready-fail")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}

	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != "DEPENDENCY_UNAVAILABLE" {
		t.Fatalf("expected code DEPENDENCY_UNAVAILABLE, got %q", body.Code)
	}
	if body.RequestID == "" {
		t.Fatalf("expected request_id")
	}
	if body.TraceID != "trace-ready-fail" {
		t.Fatalf("expected trace_id trace-ready-fail, got %q", body.TraceID)
	}
	if body.Message != "dependency unavailable" {
		t.Fatalf("expected safe failure message, got %q", body.Message)
	}
}
