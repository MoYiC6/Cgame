package bootstrap

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

func TestRequestAndTraceMiddlewarePopulateHeadersAndContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestIDMiddleware(), TraceContextMiddleware(observability.NewNoopPropagator()))
	engine.GET("/ping", func(c *gin.Context) {
		traceID, _ := observability.TraceIDFromContext(c.Request.Context())
		response.Success(c, gin.H{"trace_id": traceID})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	request.Header.Set("X-Request-ID", "req-mw")
	request.Header.Set("X-Trace-ID", "trace-mw")

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Request-ID") != "req-mw" {
		t.Fatalf("expected response request id req-mw, got %q", recorder.Header().Get("X-Request-ID"))
	}
	if recorder.Header().Get("X-Trace-ID") != "trace-mw" {
		t.Fatalf("expected response trace id trace-mw, got %q", recorder.Header().Get("X-Trace-ID"))
	}

	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.RequestID != "req-mw" {
		t.Fatalf("expected request_id req-mw, got %q", body.RequestID)
	}
	data, ok := body.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got %#v", body.Data)
	}
	if data["trace_id"] != "trace-mw" {
		t.Fatalf("expected trace_id trace-mw, got %#v", data["trace_id"])
	}
}

func TestRecoveryMiddlewareConvertsPanicToJSONError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RequestIDMiddleware(), RecoveryMiddleware(logger.New("debug", io.Discard)))
	engine.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}
}
