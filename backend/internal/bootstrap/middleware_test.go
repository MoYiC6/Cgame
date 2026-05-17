package bootstrap

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/platform/config"
	apperrors "backend/internal/platform/errors"
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

func TestCORSMiddlewareHandlesPreflightAndSimpleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(CORSMiddleware(config.CORSConfig{
		AllowedOrigins:   []string{"https://frontend.example.com"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAgeSecs:       600,
	}))
	engine.GET("/ping", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	preflight := httptest.NewRecorder()
	preflightReq := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	preflightReq.Header.Set("Origin", "https://frontend.example.com")
	preflightReq.Header.Set("Access-Control-Request-Method", "GET")
	preflightReq.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")
	engine.ServeHTTP(preflight, preflightReq)

	if preflight.Code != http.StatusNoContent {
		t.Fatalf("expected preflight 204, got %d", preflight.Code)
	}
	if preflight.Header().Get("Access-Control-Allow-Origin") != "https://frontend.example.com" {
		t.Fatalf("expected allow origin header, got %q", preflight.Header().Get("Access-Control-Allow-Origin"))
	}
	if preflight.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("expected allow credentials true, got %q", preflight.Header().Get("Access-Control-Allow-Credentials"))
	}
	if preflight.Header().Get("Access-Control-Max-Age") != "600" {
		t.Fatalf("expected max age 600, got %q", preflight.Header().Get("Access-Control-Max-Age"))
	}

	simple := httptest.NewRecorder()
	simpleReq := httptest.NewRequest(http.MethodGet, "/ping", nil)
	simpleReq.Header.Set("Origin", "https://frontend.example.com")
	engine.ServeHTTP(simple, simpleReq)

	if simple.Code != http.StatusOK {
		t.Fatalf("expected simple request 200, got %d", simple.Code)
	}
	if simple.Header().Get("Access-Control-Allow-Origin") != "https://frontend.example.com" {
		t.Fatalf("expected allow origin on simple request, got %q", simple.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestSecurityHeadersMiddlewareAddsConfiguredHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(SecurityHeadersMiddleware(config.SecurityHeadersConfig{
		FrameOptions:       "DENY",
		ContentTypeOptions: true,
		ReferrerPolicy:     "strict-origin-when-cross-origin",
	}))
	engine.GET("/ping", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if recorder.Header().Get("X-Frame-Options") != "DENY" {
		t.Fatalf("expected X-Frame-Options DENY, got %q", recorder.Header().Get("X-Frame-Options"))
	}
	if recorder.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("expected X-Content-Type-Options nosniff, got %q", recorder.Header().Get("X-Content-Type-Options"))
	}
	if recorder.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Fatalf("expected Referrer-Policy strict-origin-when-cross-origin, got %q", recorder.Header().Get("Referrer-Policy"))
	}
}

func TestRateLimitMiddlewareRejectsRequestsBeyondLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(RateLimitMiddleware(config.RateLimitConfig{
		Requests:   2,
		WindowSecs: 60,
	}))
	engine.GET("/ping", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	for attempt := range 2 {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/ping", nil)
		request.RemoteAddr = "127.0.0.1:12345"
		engine.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected request %d to pass, got %d", attempt+1, recorder.Code)
		}
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	request.RemoteAddr = "127.0.0.1:12345"
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", recorder.Code)
	}
	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != apperrors.Code(apperrors.NewAppError("RATE_LIMITED", "rate limit exceeded", http.StatusTooManyRequests)) {
		t.Fatalf("expected RATE_LIMITED code, got %q", body.Code)
	}
}

func TestRateLimiterEvictsExpiredEntries(t *testing.T) {
	limiter := &rateLimiter{entries: make(map[string]*rateLimitEntry)}
	window := 50 * time.Millisecond
	start := time.Now()

	if !limiter.allow("192.0.2.10", start, 1, window) {
		t.Fatal("expected first request to be allowed")
	}
	if len(limiter.entries) != 1 {
		t.Fatalf("expected one entry after first request, got %d", len(limiter.entries))
	}
	if !limiter.allow("198.51.100.20", start.Add(2*window), 1, window) {
		t.Fatal("expected second distinct request after window to be allowed")
	}
	if len(limiter.entries) != 1 {
		t.Fatalf("expected expired entry eviction to keep one active entry, got %d", len(limiter.entries))
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
