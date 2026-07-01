package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type memoryScheduler struct {
	mu        sync.Mutex
	registry  *JobRegistry
	ticker    *time.Ticker
	done      chan struct{}
	running   int
	wg        sync.WaitGroup
	mw        []JobMiddleware
	started   bool
}

func NewMemoryScheduler() Scheduler {
	return &memoryScheduler{
		registry: NewJobRegistry(),
		done:     make(chan struct{}),
	}
}

func (m *memoryScheduler) Register(job Job) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.started {
		return fmt.Errorf("cannot register job after scheduler started")
	}
	return m.registry.Register(job)
}

func (m *memoryScheduler) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.registry.Remove(name)
}

func (m *memoryScheduler) Use(middlewares ...JobMiddleware) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mw = append(m.mw, middlewares...)
}

func (m *memoryScheduler) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.started {
		m.mu.Unlock()
		return nil
	}
	m.started = true
		m.ticker = time.NewTicker(100 * time.Millisecond)
	m.done = make(chan struct{})
	m.mu.Unlock()

	defer m.ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.mu.Lock()
			m.done = make(chan struct{})
			closed := make(chan struct{})
			go func() {
				m.wg.Wait()
				close(closed)
			}()
			m.mu.Unlock()
			select {
			case <-closed:
			case <-time.After(5 * time.Second):
			}
			return nil
		case <-m.ticker.C:
			for _, entry := range m.registry.Due(time.Now()) {
				m.runJob(ctx, entry)
			}
		}
	}
}

func (m *memoryScheduler) runJob(ctx context.Context, entry *jobEntry) {
	m.mu.Lock()
	if !m.started {
		m.mu.Unlock()
		return
	}
	m.running++
	m.wg.Add(1)
	m.mu.Unlock()

	defer m.wg.Done()

	jobCtx := WithJobName(ctx, entry.job.Name)

	fn := entry.job.Job
	for i := len(m.mw) - 1; i >= 0; i-- {
		fn = m.mw[i](fn)
	}

	_ = fn(jobCtx)

	m.mu.Lock()
	m.running--
	m.mu.Unlock()
}

func (m *memoryScheduler) Stop(ctx context.Context) error {
	return nil
}
