package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

func TestHandlerLoginValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&stubHandlerService{}, HandlerConfig{Cookie: testCookieConfig()})
	engine := newHandlerEngine(h)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"username":""}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	assertAPIError(t, resp, http.StatusBadRequest, "INVALID_ARGUMENT", "req-handler", "trace-handler")
}

func TestHandlerLoginSuccessWritesCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&stubHandlerService{
		loginResponse: &AuthResponse{AccessToken: "access", TokenType: "Bearer", ExpiresIn: 900, Username: "admin", Nickname: "管理员"},
		loginCookie:   &RefreshCookie{Value: "refresh", ExpiresAt: time.Now().UTC().Add(time.Hour)},
	}, HandlerConfig{Cookie: testCookieConfig()})
	engine := newHandlerEngine(h)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"username":"admin@example.com","password":"secret-password"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	if cookies := resp.Result().Cookies(); len(cookies) == 0 || cookies[0].Name != "refresh_token" || cookies[0].Value != "refresh" {
		t.Fatalf("expected refresh cookie to be set, got %#v", cookies)
	}
	var body response.APIResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != "OK" {
		t.Fatalf("expected code OK, got %q", body.Code)
	}
}

func TestHandlerLoginFailureReturnsInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&stubHandlerService{loginErr: ErrInvalidCredentials}, HandlerConfig{Cookie: testCookieConfig()})
	engine := newHandlerEngine(h)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"username":"admin@example.com","password":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	assertAPIError(t, resp, http.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "req-handler", "trace-handler")
}

func TestHandlerRefreshWithoutCookieReturnsUnauthorizedAndClearsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&stubHandlerService{refreshErr: ErrRefreshInvalid}, HandlerConfig{Cookie: testCookieConfig()})
	engine := newHandlerEngine(h)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	assertAPIError(t, resp, http.StatusUnauthorized, "AUTH_REFRESH_INVALID", "req-handler", "trace-handler")
	assertClearedRefreshCookie(t, resp)
}

func TestHandlerRefreshSuccessRotatesCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&stubHandlerService{
		refreshResponse: &AuthResponse{AccessToken: "new-access", TokenType: "Bearer", ExpiresIn: 900, Username: "admin", Nickname: "管理员"},
		refreshCookie:   &RefreshCookie{Value: "new-refresh", ExpiresAt: time.Now().UTC().Add(time.Hour)},
	}, HandlerConfig{Cookie: testCookieConfig()})
	engine := newHandlerEngine(h)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old-refresh"})
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	if cookies := resp.Result().Cookies(); len(cookies) == 0 || cookies[0].Value != "new-refresh" {
		t.Fatalf("expected rotated refresh cookie, got %#v", cookies)
	}
}

func TestHandlerLogoutWithStaleCookieStillClearsCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&stubHandlerService{}, HandlerConfig{Cookie: testCookieConfig()})
	engine := newHandlerEngine(h)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "stale-refresh"})
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}
	assertClearedRefreshCookie(t, resp)
}

func TestHandlerMeUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&stubHandlerService{meErr: ErrUnauthorized}, HandlerConfig{Cookie: testCookieConfig()})
	engine := newHandlerEngine(h)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	assertAPIError(t, resp, http.StatusUnauthorized, "AUTH_UNAUTHORIZED", "req-handler", "trace-handler")
}

func TestHandlerMeAuthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewHandler(&stubHandlerService{meResponse: &MeResponse{Username: "admin", Nickname: "管理员", Roles: []string{"admin"}, Permissions: []string{}, SessionID: "ses_1"}}, HandlerConfig{Cookie: testCookieConfig()})
	engine := newHandlerEngine(h)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}
}

type stubHandlerService struct {
	loginResponse   *AuthResponse
	loginCookie     *RefreshCookie
	loginErr        error
	refreshResponse *AuthResponse
	refreshCookie   *RefreshCookie
	refreshErr      error
	logoutErr       error
	meResponse      *MeResponse
	meErr           error
	lastLoginReq    *LoginRequest
	lastRefreshReq  *RefreshRequest
	lastLogoutReq   *LogoutRequest
}

func (s *stubHandlerService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, *RefreshCookie, error) {
	s.lastLoginReq = req
	return s.loginResponse, s.loginCookie, s.loginErr
}

func (s *stubHandlerService) Refresh(ctx context.Context, req *RefreshRequest) (*AuthResponse, *RefreshCookie, error) {
	s.lastRefreshReq = req
	return s.refreshResponse, s.refreshCookie, s.refreshErr
}

func (s *stubHandlerService) Logout(ctx context.Context, req *LogoutRequest) error {
	s.lastLogoutReq = req
	return s.logoutErr
}

func (s *stubHandlerService) Me(ctx context.Context) (*MeResponse, error) {
	return s.meResponse, s.meErr
}

func newHandlerEngine(h *Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(withObservability("req-handler", "trace-handler"))
	api := engine.Group("/api")
	h.RegisterRoutes(api)
	return engine
}

func assertClearedRefreshCookie(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()
	cookies := recorder.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected at least one cookie")
	}
	if cookies[0].Name != "refresh_token" {
		t.Fatalf("expected refresh_token cookie, got %#v", cookies[0])
	}
	if cookies[0].MaxAge != -1 && cookies[0].Value != "" {
		t.Fatalf("expected cleared cookie, got %#v", cookies[0])
	}
}

func testCookieConfig() CookieConfig {
	return CookieConfig{Name: "refresh_token", Path: "/api/auth", HTTPOnly: true, SameSite: "lax"}
}
