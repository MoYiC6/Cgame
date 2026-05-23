package bootstrap

import (
	"context"
	"io"
	"testing"
	"time"

	"backend/internal/modules/user"
	"backend/internal/platform/logger"
	"backend/internal/platform/security"
)

type stubTask struct{}

type contextCaptureTask struct {
	principal *security.Principal
	afterRun  func()
}

func (stubTask) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (stubTask) Probe(ctx context.Context) error {
	return nil
}

func (t *contextCaptureTask) Run(ctx context.Context) error {
	principal, ok := security.PrincipalFromContext(ctx)
	if !ok {
		return context.Canceled
	}
	t.principal = principal
	if t.afterRun != nil {
		t.afterRun()
	}
	return context.Canceled
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

func TestRegisterRunnableWithSystemPrincipalInjectsWorkerIdentity(t *testing.T) {
	worker := NewWorker(logger.New("debug", io.Discard))
	ctx, cancel := context.WithCancel(context.Background())
	task := &contextCaptureTask{afterRun: cancel}

	worker.RegisterRunnableWithSystemPrincipal("sync-orders", task)

	err := worker.Run(ctx)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if task.principal == nil {
		t.Fatal("expected principal to be injected")
	}
	if task.principal.UserID != "worker:sync-orders" {
		t.Fatalf("expected worker user id, got %+v", task.principal)
	}
	if task.principal.SessionID != "worker:sync-orders" {
		t.Fatalf("expected worker session id, got %+v", task.principal)
	}
	if task.principal.Status != user.StatusActive {
		t.Fatalf("expected active worker principal, got %+v", task.principal)
	}
	if !security.HasRole(task.principal, "system") {
		t.Fatalf("expected system role, got %+v", task.principal)
	}
	if !security.HasPermission(task.principal, "internal:worker") {
		t.Fatalf("expected internal worker permission, got %+v", task.principal)
	}
}
