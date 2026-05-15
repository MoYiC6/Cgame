package integration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/response"
)

type stubDB struct {
	pingErr error
}

var _ interface{ Ping(context.Context) error } = stubDB{}

func (s stubDB) Ping(ctx context.Context) error {
	return s.pingErr
}

type recordingReadyDB struct {
	db        interface{ Ping(context.Context) error }
	pingCalls int
}

func (d *recordingReadyDB) Ping(ctx context.Context) error {
	if d == nil || d.db == nil {
		return errors.New("db is nil")
	}
	d.pingCalls++
	return d.db.Ping(ctx)
}

func TestReadyzReturnsOKWhenDBAvailable(t *testing.T) {
	recording := &recordingReadyDB{db: stubDB{}}
	engine := newIntegrationEngine(recording)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	request.Header.Set("X-Trace-ID", "trace-ready-healthy")
	request.Header.Set("X-Request-ID", "req-ready-healthy")

	engine.ServeHTTP(recorder, request)

	if recording.pingCalls != 1 {
		t.Fatalf("expected readiness to call DB ping once, got %d", recording.pingCalls)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Trace-ID") != "trace-ready-healthy" {
		t.Fatalf("expected X-Trace-ID trace-ready-healthy, got %q", recorder.Header().Get("X-Trace-ID"))
	}
	if recorder.Header().Get("X-Request-ID") != "req-ready-healthy" {
		t.Fatalf("expected X-Request-ID req-ready-healthy, got %q", recorder.Header().Get("X-Request-ID"))
	}

	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != "OK" {
		t.Fatalf("expected code OK, got %q", body.Code)
	}
	if body.RequestID != "req-ready-healthy" {
		t.Fatalf("expected request_id req-ready-healthy, got %q", body.RequestID)
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
}

func TestReadyzFailsWhenDBUnavailable(t *testing.T) {
	engine := newIntegrationEngine(stubDB{pingErr: errors.New("db down")})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	request.Header.Set("X-Trace-ID", "trace-ready-fail")
	request.Header.Set("X-Request-ID", "req-ready-fail")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Trace-ID") != "trace-ready-fail" {
		t.Fatalf("expected X-Trace-ID trace-ready-fail, got %q", recorder.Header().Get("X-Trace-ID"))
	}
	if recorder.Header().Get("X-Request-ID") != "req-ready-fail" {
		t.Fatalf("expected X-Request-ID req-ready-fail, got %q", recorder.Header().Get("X-Request-ID"))
	}

	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != "DEPENDENCY_UNAVAILABLE" {
		t.Fatalf("expected code DEPENDENCY_UNAVAILABLE, got %q", body.Code)
	}
	if body.Message != "dependency unavailable" {
		t.Fatalf("expected safe failure message, got %q", body.Message)
	}
	if body.RequestID != "req-ready-fail" {
		t.Fatalf("expected request_id req-ready-fail, got %q", body.RequestID)
	}
	if body.TraceID != "trace-ready-fail" {
		t.Fatalf("expected trace_id trace-ready-fail, got %q", body.TraceID)
	}
}
