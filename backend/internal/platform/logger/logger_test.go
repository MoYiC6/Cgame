package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"backend/internal/platform/config"
	"backend/internal/platform/observability"
)

func TestWithContextAddsRequestAndTraceIDs(t *testing.T) {
	var buffer bytes.Buffer
	base := NewText("debug", &buffer)
	ctx := context.Background()
	ctx = observability.WithRequestID(ctx, "req-log")
	ctx = observability.WithTraceID(ctx, "trace-log")

	WithContext(ctx, base).Info("boot ok", "component", "api")

	output := buffer.String()
	if !strings.Contains(output, "req-log") {
		t.Fatalf("expected output to contain request id, got %s", output)
	}
	if !strings.Contains(output, "trace-log") {
		t.Fatalf("expected output to contain trace id, got %s", output)
	}
}

func TestSamplingRateZeroSuppressesDebugInfo(t *testing.T) {
	var buffer bytes.Buffer
	log := NewWithSample("debug", "text", 0.0, &buffer)

	log.Debug("debug-msg")
	log.Info("info-msg")
	log.Warn("warn-msg")
	log.Error("error-msg")

	output := buffer.String()
	if strings.Contains(output, "debug-msg") {
		t.Fatalf("sample rate 0 should suppress debug, got: %s", output)
	}
	if strings.Contains(output, "info-msg") {
		t.Fatalf("sample rate 0 should suppress info, got: %s", output)
	}
	if !strings.Contains(output, "warn-msg") {
		t.Fatalf("warn should always appear, got: %s", output)
	}
	if !strings.Contains(output, "error-msg") {
		t.Fatalf("error should always appear, got: %s", output)
	}
}

func TestSamplingRateOneLogsAll(t *testing.T) {
	var buffer bytes.Buffer
	log := NewWithSample("debug", "text", 1.0, &buffer)

	log.Debug("debug-msg")
	log.Info("info-msg")

	output := buffer.String()
	if !strings.Contains(output, "debug-msg") {
		t.Fatalf("sample rate 1.0 should log all debug, got: %s", output)
	}
	if !strings.Contains(output, "info-msg") {
		t.Fatalf("sample rate 1.0 should log all info, got: %s", output)
	}
}

func TestSamplingRateConfig(t *testing.T) {
	cfg := config.LogConfig{Level: "debug", Format: "text", SampleRate: 0.0}
	if cfg.SampleRate != 0.0 {
		t.Fatalf("expected SampleRate 0.0, got %f", cfg.SampleRate)
	}
	_ = New(cfg)
}
