package scheduler_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"backend/internal/platform/scheduler"
)

func TestMemorySchedulerRegistersAndStarts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := scheduler.NewMemoryScheduler()
	var mu sync.Mutex
	var count int

	if err := s.Register(scheduler.Job{
		Name:     "tick",
		Schedule: "100ms",
		Job: func(ctx context.Context) error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	done := make(chan struct{})
	go func() {
		_ = s.Start(ctx)
		close(done)
	}()

	<-time.After(350 * time.Millisecond)
	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if count < 2 {
		t.Fatalf("job executed %d times, want >= 2", count)
	}
}

func TestMemorySchedulerRemove(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := scheduler.NewMemoryScheduler()
	var mu sync.Mutex
	var count int

	if err := s.Register(scheduler.Job{
		Name:     "tick",
		Schedule: "100ms",
		Job: func(ctx context.Context) error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	done := make(chan struct{})
	go func() {
		_ = s.Start(ctx)
		close(done)
	}()

	<-time.After(150 * time.Millisecond)

	if err := s.Remove("tick"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	<-time.After(300 * time.Millisecond)
	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if count < 1 {
		t.Fatalf("job executed %d times, want >= 1", count)
	}
}

func TestMemorySchedulerDuplicateRegister(t *testing.T) {
	s := scheduler.NewMemoryScheduler()

	if err := s.Register(scheduler.Job{
		Name:     "dup",
		Schedule: "1s",
		Job:      func(ctx context.Context) error { return nil },
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if err := s.Register(scheduler.Job{
		Name:     "dup",
		Schedule: "1s",
		Job:      func(ctx context.Context) error { return nil },
	}); err == nil {
		t.Fatal("duplicate Register() should return error")
	}
}

func TestMemorySchedulerCronSchedule(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := scheduler.NewMemoryScheduler()
	var mu sync.Mutex
	var count int

	if err := s.Register(scheduler.Job{
		Name:     "cron",
		Schedule: "*/2 * * * * *",
		Job: func(ctx context.Context) error {
			mu.Lock()
			count++
			mu.Unlock()
			return nil
		},
	}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	done := make(chan struct{})
	go func() {
		_ = s.Start(ctx)
		close(done)
	}()

	<-time.After(3 * time.Second)
	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()
	if count < 1 {
		t.Fatalf("cron job executed %d times, want >= 1", count)
	}
}
