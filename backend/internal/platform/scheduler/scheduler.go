package scheduler

import (
	"context"
	"time"
)

type JobFunc func(ctx context.Context) error

type Job struct {
	Name     string
	Schedule string
	Job      JobFunc
}

type RetryPolicy struct {
	MaxRetries int
	Backoff    time.Duration
}

type JobMiddleware func(next JobFunc) JobFunc

type Scheduler interface {
	Register(job Job) error
	Remove(name string) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Use(middlewares ...JobMiddleware)
}
