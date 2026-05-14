package bootstrap

import (
	"context"
	"io"
	"testing"
	"time"

	"backend/internal/platform/logger"
)

type stubTask struct{}

func (stubTask) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (stubTask) Probe(ctx context.Context) error {
	return nil
}

func TestWorkerRunStopsOnContextCancel(t *testing.T) {
	worker := NewWorker(logger.New("debug", io.Discard))
	worker.RegisterTask("placeholder", func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- worker.Run(ctx)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run() did not return after context cancel")
	}
}

func TestWorkerProbeRecognizesOptionalTaskProbe(t *testing.T) {
	worker := NewWorker(logger.New("debug", io.Discard))
	worker.RegisterRunnable("probeable", stubTask{})

	if err := worker.Probe(context.Background()); err != nil {
		t.Fatalf("Probe() error = %v, want nil", err)
	}
}

func TestWorkerShutdownHandlesFailures(t *testing.T) {
	worker := NewWorker(logger.New("debug", io.Discard))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := worker.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v, want nil", err)
	}
}
