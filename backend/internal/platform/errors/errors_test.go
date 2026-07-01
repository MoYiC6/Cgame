package apperrors

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"testing"
)

func TestWithCausePreservesMetadata(t *testing.T) {
	cause := stderrors.New("database down")
	err := NewAppError("INTERNAL_ERROR", "internal error", http.StatusInternalServerError)
	err.WithCause(cause)

	if err.Code != "INTERNAL_ERROR" {
		t.Fatalf("expected code INTERNAL_ERROR, got %q", err.Code)
	}
	if err.HTTPStatus != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", err.HTTPStatus)
	}
	if err.Cause != cause {
		t.Fatal("expected cause to be set by WithCause")
	}
	if !stderrors.Is(err, cause) {
		t.Fatal("expected errors.Is to match cause")
	}
}

func TestMetadataHelpersHandleGenericErrors(t *testing.T) {
	err := stderrors.New("boom")
	if Code(err) != CodeInternal {
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

func TestStackCapture(t *testing.T) {
	err := New("TEST", "test", http.StatusInternalServerError, nil)

	if len(err.stack) == 0 {
		t.Fatal("expected stack to be captured")
	}

	fn := runtime.FuncForPC(err.stack[0])
	if fn == nil {
		t.Fatal("expected to resolve first stack frame")
	}
	if !strings.Contains(fn.Name(), "TestStackCapture") {
		t.Fatalf("expected first frame to be TestStackCapture, got %q", fn.Name())
	}
}

func TestFormatVerboseIncludesStack(t *testing.T) {
	cause := stderrors.New("root cause")
	err := New("TEST", "test error", http.StatusInternalServerError, cause)

	formatted := fmt.Sprintf("%+v", err)

	if !strings.Contains(formatted, "test error") {
		t.Fatal("expected formatted output to contain message")
	}
	if !strings.Contains(formatted, "root cause") {
		t.Fatal("expected formatted output to contain cause")
	}
	if !strings.Contains(formatted, ".go:") {
		t.Fatal("expected formatted output to contain file:line from stack trace")
	}
}

func TestFormatSimpleDoesNotIncludeStack(t *testing.T) {
	err := New("TEST", "test error", http.StatusInternalServerError, nil)

	formatted := fmt.Sprintf("%v", err)

	if formatted != "test error" {
		t.Fatalf("expected simple format to be message only, got %q", formatted)
	}
}

func TestFormatQuote(t *testing.T) {
	err := New("TEST", "test error", http.StatusInternalServerError, nil)

	formatted := fmt.Sprintf("%q", err)

	if formatted != `"test error"` {
		t.Fatalf("expected quoted message, got %q", formatted)
	}
}

func TestStackTraceReturnsNilForNonAppError(t *testing.T) {
	err := stderrors.New("plain error")
	stack := StackTrace(err)
	if stack != nil {
		t.Fatal("expected nil stack for non-AppError")
	}
}

func TestStackTraceReturnsStackForAppError(t *testing.T) {
	err := New("TEST", "test", http.StatusInternalServerError, nil)
	stack := StackTrace(err)
	if len(stack) == 0 {
		t.Fatal("expected non-empty stack for AppError")
	}
}

func TestCodeFromConstants(t *testing.T) {
	if CodeOK != "OK" {
		t.Fatalf("expected CodeOK to be OK, got %q", CodeOK)
	}
	if CodeInternal != "INTERNAL_ERROR" {
		t.Fatalf("expected CodeInternal to be INTERNAL_ERROR, got %q", CodeInternal)
	}
}

func TestNilSafety(t *testing.T) {
	var nilErr *AppError
	if nilErr.Error() != "internal error" {
		t.Fatalf("expected nil-safe Error()")
	}
	if nilErr.Unwrap() != nil {
		t.Fatalf("expected nil-safe Unwrap()")
	}
	nilErr.WithCause(stderrors.New("x"))
	if nilErr != nil {
		t.Fatal("expected WithCause on nil to stay nil")
	}
}
