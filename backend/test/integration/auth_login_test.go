package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/platform/database"
	"backend/internal/platform/response"
)

func TestAuthLoginValidationError(t *testing.T) {
	engine := newIntegrationEngine(database.DummyDB{})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"identifier":""}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", resp.Code, resp.Body.String())
	}
	var body response.APIResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != "INVALID_ARGUMENT" {
		t.Fatalf("expected INVALID_ARGUMENT, got %q", body.Code)
	}
}
