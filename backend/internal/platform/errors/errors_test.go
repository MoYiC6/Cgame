package apperrors

import (
	stderrors "errors"
	"net/http"
	"testing"
)

func TestWrapPreservesMetadata(t *testing.T) {
	base := NewAppError("INTERNAL_ERROR", "internal error", http.StatusInternalServerError)
	cause := stderrors.New("database down")
	wrapped := Wrap(base, cause)

	if wrapped.Code != "INTERNAL_ERROR" {
		t.Fatalf("expected code INTERNAL_ERROR, got %q", wrapped.Code)
	}
	if wrapped.HTTPStatus != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", wrapped.HTTPStatus)
	}
	if !stderrors.Is(wrapped, cause) {
		t.Fatal("expected wrapped error to expose its cause through errors.Is")
	}
}

func TestMetadataHelpersHandleGenericErrors(t *testing.T) {
	err := stderrors.New("boom")
	if Code(err) != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR fallback, got %q", Code(err))
	}
	if Status(err) != http.StatusInternalServerError {
		t.Fatalf("expected 500 fallback, got %d", Status(err))
	}
	if SafeMessage(err) != "internal error" {
		t.Fatalf("expected generic safe message, got %q", SafeMessage(err))
	}
}

func TestNewSetsCause(t *testing.T) {
	cause := stderrors.New("redis unavailable")
	err := New("REDIS_UNAVAILABLE", "redis unavailable", http.StatusServiceUnavailable, cause)

	if err.Cause != cause {
		t.Fatal("expected cause to be set by New")
	}
	if !stderrors.Is(err, cause) {
		t.Fatal("expected errors.Is to match cause")
	}
}
