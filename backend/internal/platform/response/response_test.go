package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/observability"
	"github.com/gin-gonic/gin"
)

func TestSuccessWritesRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req-123"))
	ctx.Request = request

	Success(ctx, gin.H{"module": "health"})

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var body APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.RequestID != "req-123" {
		t.Fatalf("expected request_id req-123, got %q", body.RequestID)
	}
	if body.Code != "OK" {
		t.Fatalf("expected code OK, got %q", body.Code)
	}
}

func TestFailUsesAppErrorMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	request = request.WithContext(observability.WithRequestID(request.Context(), "req-500"))
	ctx.Request = request

	Fail(ctx, apperrors.NewAppError("INVALID_ARGUMENT", "invalid input", http.StatusBadRequest))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}

	var body APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != "INVALID_ARGUMENT" {
		t.Fatalf("expected INVALID_ARGUMENT, got %q", body.Code)
	}
	if body.Message != "invalid input" {
		t.Fatalf("expected invalid input, got %q", body.Message)
	}
	if body.RequestID != "req-500" {
		t.Fatalf("expected request_id req-500, got %q", body.RequestID)
	}
}

func TestFailFallsBackForGenericError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	ctx.Request = request

	Fail(ctx, assertErr("boom"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}

	var body APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	if body.Code != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR, got %q", body.Code)
	}
	if body.Message != "internal error" {
		t.Fatalf("expected internal error, got %q", body.Message)
	}
}

type assertErr string

func (e assertErr) Error() string { return string(e) }
