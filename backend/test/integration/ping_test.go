package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/bootstrap"
	"backend/internal/modules/inventory"
	"backend/internal/modules/notification"
	"backend/internal/modules/order"
	"backend/internal/modules/payment"
	"backend/internal/platform/config"
	"backend/internal/platform/database"
	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

func TestBootstrapRegistersAllPingRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := bootstrap.Dependencies{
		Config: config.Config{
			App:    config.AppConfig{Name: "backend-test", Env: "test"},
			Server: config.ServerConfig{Addr: ":18080"},
			Log:    config.LogConfig{Level: "debug"},
		},
		Logger:     logger.New("debug", io.Discard),
		Tracer:     observability.NewNoopTracer(),
		Propagator: observability.NewNoopPropagator(),
	}

	engine := bootstrap.NewAPIEngine(
		deps,
		order.NewHandler(order.NewService(order.NewRepository(), database.NoopTxManager{})),
		payment.NewHandler(payment.NewService(payment.NewRepository(), database.NoopTxManager{})),
		inventory.NewHandler(inventory.NewService(inventory.NewRepository(), database.NoopTxManager{})),
		notification.NewHandler(notification.NewService(notification.NewRepository(), database.NoopTxManager{})),
	)

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

			engine.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200 for %s, got %d", tt.path, recorder.Code)
			}
			if recorder.Header().Get("X-Request-ID") == "" {
				t.Fatalf("expected X-Request-ID header for %s", tt.path)
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
			if body.TraceID != "trace-int" {
				t.Fatalf("expected top-level trace_id trace-int, got %q", body.TraceID)
			}
		})
	}
}
