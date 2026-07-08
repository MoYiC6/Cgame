package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/database"
	"backend/internal/platform/response"
)

func TestAuthRefreshWithoutCookieReturnsUnauthorized(t *testing.T) {
	engine := newIntegrationEngine(database.DummyDB{})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	resp := httptest.NewRecorder()

	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", resp.Code, resp.Body.String())
	}
	var body response.APIResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != "AUTH_REFRESH_INVALID" {
		t.Fatalf("expected AUTH_REFRESH_INVALID, got %q", body.Code)
	}
}
