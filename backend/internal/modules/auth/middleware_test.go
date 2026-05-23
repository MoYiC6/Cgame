package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"backend/internal/platform/security"
	"github.com/gin-gonic/gin"
)

func TestAuthMiddlewareMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		ctx := observability.WithRequestID(c.Request.Context(), "req-missing")
		ctx = observability.WithTraceID(ctx, "trace-missing")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	engine.Use(AuthMiddleware(newTestTokenManager(t)))
	engine.GET("/protected", func(c *gin.Context) { response.Success(c, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	assertAPIError(t, resp, http.StatusUnauthorized, "AUTH_TOKEN_MISSING", "req-missing", "trace-missing")
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(withObservability("req-invalid", "trace-invalid"))
	engine.Use(AuthMiddleware(newTestTokenManager(t)))
	engine.GET("/protected", func(c *gin.Context) { response.Success(c, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not-a-token")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	assertAPIError(t, resp, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "req-invalid", "trace-invalid")
}

func TestAuthMiddlewareExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mgr := security.NewHMACTokenManager(security.HMACTokenConfig{
		Issuer:         "backend",
		Audience:       "admin-api",
		KeyID:          "test-key",
		Secret:         []byte("01234567890123456789012345678901"),
		AccessTokenTTL: -1 * time.Minute,
		ClockSkew:      0,
	})
	tok, err := mgr.IssueAccessToken(t.Context(), &security.Principal{PublicID: "usr_1", SessionID: "ses_1", Roles: []string{"admin"}, Permissions: []string{"order:read"}})
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	engine := gin.New()
	engine.Use(withObservability("req-expired", "trace-expired"))
	engine.Use(AuthMiddleware(newTestTokenManager(t)))
	engine.GET("/protected", func(c *gin.Context) { response.Success(c, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok.Token)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	assertAPIError(t, resp, http.StatusUnauthorized, "AUTH_TOKEN_EXPIRED", "req-expired", "trace-expired")
}

func TestAuthMiddlewareStoresPrincipal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mgr := newTestTokenManager(t)
	tok, err := mgr.IssueAccessToken(t.Context(), &security.Principal{PublicID: "usr_1", SessionID: "ses_1", Roles: []string{"admin"}, Permissions: []string{"order:read", "order:read"}})
	if err != nil {
		t.Fatalf("IssueAccessToken() error = %v", err)
	}

	engine := gin.New()
	engine.Use(AuthMiddleware(mgr))
	engine.GET("/protected", func(c *gin.Context) {
		principal, ok := security.PrincipalFromContext(c.Request.Context())
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "missing principal"})
			return
		}
		sessionID, ok := security.SessionIDFromContext(c.Request.Context())
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "missing session id"})
			return
		}
		response.Success(c, gin.H{"public_id": principal.PublicID, "session_id": sessionID, "permissions": principal.Permissions})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tok.Token)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func TestRequirePermissionForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		ctx := security.WithPrincipal(c.Request.Context(), &security.Principal{PublicID: "usr_1", SessionID: "ses_1", Roles: []string{"admin"}, Permissions: []string{"order:read"}})
		ctx = observability.WithRequestID(ctx, "req-forbidden")
		ctx = observability.WithTraceID(ctx, "trace-forbidden")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	engine.GET("/protected", RequirePermission("order:cancel"), func(c *gin.Context) { response.Success(c, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	assertAPIError(t, resp, http.StatusForbidden, "AUTH_FORBIDDEN", "req-forbidden", "trace-forbidden")
}

func TestRequirePermissionAllowsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		ctx := security.WithPrincipal(c.Request.Context(), &security.Principal{PublicID: "usr_1", SessionID: "ses_1", Roles: []string{"admin"}, Permissions: []string{"order:read"}})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	engine.GET("/protected", RequirePermission("order:read"), func(c *gin.Context) { response.Success(c, gin.H{"ok": true}) })

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}
}

func withObservability(requestID, traceID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := observability.WithRequestID(c.Request.Context(), requestID)
		ctx = observability.WithTraceID(ctx, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func newTestTokenManager(t *testing.T) security.TokenManager {
	t.Helper()
	return security.NewHMACTokenManager(security.HMACTokenConfig{
		Issuer:         "backend",
		Audience:       "admin-api",
		KeyID:          "test-key",
		Secret:         []byte("01234567890123456789012345678901"),
		AccessTokenTTL: 15 * time.Minute,
		ClockSkew:      30 * time.Second,
	})
}

func assertAPIError(t *testing.T, recorder *httptest.ResponseRecorder, wantStatus int, wantCode, wantRequestID, wantTraceID string) {
	t.Helper()
	if recorder.Code != wantStatus {
		t.Fatalf("expected status %d, got %d body=%s", wantStatus, recorder.Code, recorder.Body.String())
	}
	var body response.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != wantCode {
		t.Fatalf("expected code %q, got %q", wantCode, body.Code)
	}
	if body.RequestID != wantRequestID {
		t.Fatalf("expected request_id %q, got %q", wantRequestID, body.RequestID)
	}
	if body.TraceID != wantTraceID {
		t.Fatalf("expected trace_id %q, got %q", wantTraceID, body.TraceID)
	}
	if apperrors.Code(apperrors.NewAppError(body.Code, body.Message, wantStatus)) == "" {
		t.Fatal("expected app error code to be non-empty")
	}
}
