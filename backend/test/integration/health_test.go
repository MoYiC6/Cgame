package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/response"
)

func TestHealthzReturnsOKAndCarriesTraceAndRequestIDs(t *testing.T) {
	engine := newIntegrationEngine(nil)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("X-Trace-ID", "trace-health")
	request.Header.Set("X-Request-ID", "req-health")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Trace-ID") != "trace-health" {
		t.Fatalf("expected X-Trace-ID trace-health, got %q", recorder.Header().Get("X-Trace-ID"))
	}
	if recorder.Header().Get("X-Request-ID") != "req-health" {
		t.Fatalf("expected X-Request-ID req-health, got %q", recorder.Header().Get("X-Request-ID"))
	}

	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != "OK" {
		t.Fatalf("expected code OK, got %q", body.Code)
	}
	if body.RequestID != "req-health" {
		t.Fatalf("expected request_id req-health, got %q", body.RequestID)
	}
	if body.TraceID != "trace-health" {
		t.Fatalf("expected trace_id trace-health, got %q", body.TraceID)
	}
	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got %#v", body.Data)
	}
	if data["status"] != "ok" {
		t.Fatalf("expected status ok, got %#v", data["status"])
	}
}
