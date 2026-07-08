package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/database"
	"backend/internal/platform/response"
)

func TestAuthMeRouteIsRegistered(t *testing.T) {
	engine := newIntegrationEngine(database.DummyDB{})
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	resp := httptest.NewRecorder()

	engine.ServeHTTP(resp, req)

	if resp.Code == http.StatusNotFound {
		t.Fatalf("expected auth route to exist, got 404")
	}
}

func TestAuthMeWithoutBearerReturnsUnauthorized(t *testing.T) {
	engine := newIntegrationEngine(database.DummyDB{})
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	resp := httptest.NewRecorder()

	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", resp.Code, resp.Body.String())
	}
	var body response.APIResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != "AUTH_TOKEN_MISSING" {
		t.Fatalf("expected AUTH_TOKEN_MISSING, got %q", body.Code)
	}
}
