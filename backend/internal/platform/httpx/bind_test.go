package httpx

import (
	"bytes"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apperrors "backend/internal/platform/errors"
	"github.com/gin-gonic/gin"
)

type bindRequest struct {
	ResourceID string `uri:"resource_id" binding:"required"`
	Page       int    `form:"page" binding:"required,min=1"`
	TenantID   string `header:"X-Tenant-ID" binding:"required"`
	Name       string `json:"name" binding:"required"`
}

func TestBindAndValidateBindsAllSupportedSources(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx := newBindTestContext(t, `/resources/42?page=2`, `{"name":"widget"}`)
	ctx.Request.Header.Set("X-Tenant-ID", "tenant-1")
	ctx.Params = gin.Params{{Key: "resource_id", Value: "42"}}

	payload, err := BindAndValidate[bindRequest](ctx)
	if err != nil {
		t.Fatalf("BindAndValidate returned error: %v", err)
	}
	if payload == nil {
		t.Fatal("expected payload, got nil")
	}
	if payload.ResourceID != "42" {
		t.Fatalf("expected ResourceID 42, got %q", payload.ResourceID)
	}
	if payload.Page != 2 {
		t.Fatalf("expected Page 2, got %d", payload.Page)
	}
	if payload.TenantID != "tenant-1" {
		t.Fatalf("expected TenantID tenant-1, got %q", payload.TenantID)
	}
	if payload.Name != "widget" {
		t.Fatalf("expected Name widget, got %q", payload.Name)
	}
}

func TestBindAndValidateMapsValidationErrorsToAppError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx := newBindTestContext(t, `/resources/42?page=2`, `{"name":"widget"}`)
	ctx.Params = gin.Params{{Key: "resource_id", Value: "42"}}

	payload, err := BindAndValidate[bindRequest](ctx)
	if payload != nil {
		t.Fatalf("expected nil payload on validation error, got %+v", payload)
	}
	assertInvalidArgumentError(t, err)
}

func TestBindAndValidateMapsBindErrorsToAppError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctx := newBindTestContext(t, `/resources/42?page=2`, `{"name":`)
	ctx.Request.Header.Set("X-Tenant-ID", "tenant-1")
	ctx.Params = gin.Params{{Key: "resource_id", Value: "42"}}

	payload, err := BindAndValidate[bindRequest](ctx)
	if payload != nil {
		t.Fatalf("expected nil payload on bind error, got %+v", payload)
	}
	assertInvalidArgumentError(t, err)
}

func newBindTestContext(t *testing.T, target string, body string) *gin.Context {
	t.Helper()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodPost, target, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	ctx.Request = request
	return ctx
}

func assertInvalidArgumentError(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var appErr *apperrors.AppError
	if !stderrors.As(err, &appErr) {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.Code != "INVALID_ARGUMENT" {
		t.Fatalf("expected code INVALID_ARGUMENT, got %q", appErr.Code)
	}
	if appErr.HTTPStatus != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", appErr.HTTPStatus)
	}
	if appErr.Cause == nil {
		t.Fatal("expected cause to be preserved")
	}
}
