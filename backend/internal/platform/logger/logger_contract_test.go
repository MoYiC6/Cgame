package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerInfoWritesStructuredFields(t *testing.T) {
	var buffer bytes.Buffer
	var log Logger = New("debug", &buffer)

	log.Info("api starting", String("addr", ":8080"), Any("component", "api"))

	output := buffer.String()
	if strings.HasPrefix(strings.TrimSpace(output), "{") {
		t.Fatalf("expected text handler output, got JSON-like output %s", output)
	}
	if !strings.Contains(output, "api starting") {
		t.Fatalf("expected output to contain message, got %s", output)
	}
	if !strings.Contains(output, ":8080") {
		t.Fatalf("expected output to contain addr field, got %s", output)
	}
	if !strings.Contains(output, "component") {
		t.Fatalf("expected output to contain component field, got %s", output)
	}
}
