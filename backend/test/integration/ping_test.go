package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/response"
)

func TestBootstrapRegistersAllPingRoutes(t *testing.T) {
	engine := newIntegrationEngine(nil)

	tests := []struct {
		name   string
		path   string
		module string
	}{
		{name: "order", path: "/api/v1/order/ping", module: "order"},
		{name: "payment", path: "/api/v1/payment/ping", module: "payment"},
		{name: "inventory", path: "/api/v1/inventory/ping", module: "inventory"},
		{name: "notification", path: "/api/v1/notification/ping", module: "notification"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			request.Header.Set("X-Trace-ID", "trace-int")
			request.Header.Set("X-Request-ID", "req-int")

			engine.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200 for %s, got %d", tt.path, recorder.Code)
			}
			if recorder.Header().Get("X-Request-ID") != "req-int" {
				t.Fatalf("expected X-Request-ID req-int for %s, got %q", tt.path, recorder.Header().Get("X-Request-ID"))
			}
			if recorder.Header().Get("X-Trace-ID") != "trace-int" {
				t.Fatalf("expected X-Trace-ID trace-int for %s, got %q", tt.path, recorder.Header().Get("X-Trace-ID"))
			}

			var body response.APIResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("json.Unmarshal returned error: %v", err)
			}
			data, ok := body.Data.(map[string]any)
			if !ok {
				t.Fatalf("expected data map for %s, got %#v", tt.path, body.Data)
			}
			if data["module"] != tt.module {
				t.Fatalf("expected module %s, got %#v", tt.module, data["module"])
			}
			if body.RequestID != "req-int" {
				t.Fatalf("expected top-level request_id req-int, got %q", body.RequestID)
			}
			if body.TraceID != "trace-int" {
				t.Fatalf("expected top-level trace_id trace-int, got %q", body.TraceID)
			}
		})
	}
}
