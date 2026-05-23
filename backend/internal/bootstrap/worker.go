package bootstrap

import (
	"context"
	"fmt"
	"sync"

	"backend/internal/modules/user"
	"backend/internal/platform/logger"
	"backend/internal/platform/security"
)

type Worker interface {
	RegisterTask(name string, handler func(ctx context.Context) error)
	Run(ctx context.Context) error
}

type RunnableTask interface {
	Run(ctx context.Context) error
}

type TaskProbe interface {
	Probe(ctx context.Context) error
}

type InMemoryWorker struct {
	log    logger.Logger
	mu     sync.Mutex
	tasks  map[string]func(ctx context.Context) error
	probes map[string]TaskProbe
}

func NewWorker(log logger.Logger) *InMemoryWorker {
	return &InMemoryWorker{
		log:    log,
		tasks:  make(map[string]func(ctx context.Context) error),
		probes: make(map[string]TaskProbe),
	}
}

func (w *InMemoryWorker) Shutdown(ctx context.Context) error {
	return nil
}

func (w *InMemoryWorker) RegisterTask(name string, handler func(ctx context.Context) error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.tasks[name] = handler
}

func (w *InMemoryWorker) RegisterRunnable(name string, task RunnableTask) {
	w.RegisterTask(name, task.Run)
	w.registerProbe(name, task)
}

func (w *InMemoryWorker) RegisterRunnableWithSystemPrincipal(name string, task RunnableTask) {
	principal := &security.Principal{
		UserID:      "worker:" + name,
		SessionID:   "worker:" + name,
		Roles:       []string{"system"},
		Permissions: []string{"internal:worker"},
		Status:      user.StatusActive,
	}
	w.RegisterTask(name, func(ctx context.Context) error {
		ctx = security.WithPrincipal(ctx, principal)
		ctx = security.WithSessionID(ctx, principal.SessionID)
		return task.Run(ctx)
	})
	w.registerProbe(name, task)
}

func (w *InMemoryWorker) registerProbe(name string, task RunnableTask) {
	probe, ok := task.(TaskProbe)
	if !ok {
		return
	}
	w.mu.Lock()
	w.probes[name] = probe
	w.mu.Unlock()
}

func (w *InMemoryWorker) Probe(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	for name, probe := range w.probes {
		if err := probe.Probe(ctx); err != nil {
			return fmt.Errorf("probe %s: %w", name, err)
		}
	}
	return nil
}

func (w *InMemoryWorker) Run(ctx context.Context) error {
	w.mu.Lock()
	copied := make(map[string]func(ctx context.Context) error, len(w.tasks))
	for name, task := range w.tasks {
		copied[name] = task
	}
	w.mu.Unlock()

	var wg sync.WaitGroup
	errCh := make(chan error, len(copied))
	for name, task := range copied {
		wg.Add(1)
		go func(name string, run func(ctx context.Context) error) {
			defer wg.Done()
			if err := run(ctx); err != nil && err != context.Canceled {
				errCh <- fmt.Errorf("task %s: %w", name, err)
			}
		}(name, task)
	}

	<-ctx.Done()
	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
