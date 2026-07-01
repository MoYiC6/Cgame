package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"
)

func LoggingMiddleware(log *slog.Logger) JobMiddleware {
	return func(next JobFunc) JobFunc {
		return func(ctx context.Context) error {
			start := time.Now()
			log.Info("job started", "job", getJobName(ctx))
			err := next(ctx)
			log.Info("job finished", "job", getJobName(ctx), "duration_ms", time.Since(start).Milliseconds(), "error", err)
			return err
		}
	}
}

func RecoveryMiddleware(log *slog.Logger) JobMiddleware {
	return func(next JobFunc) JobFunc {
		return func(ctx context.Context) error {
			defer func() {
				if r := recover(); r != nil {
					log.Error("job panic recovered", "job", getJobName(ctx), "panic", r, "stack", string(debug.Stack()))
				}
			}()
			return next(ctx)
		}
	}
}

func TimeoutMiddleware(timeout time.Duration) JobMiddleware {
	return func(next JobFunc) JobFunc {
		return func(ctx context.Context) error {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return next(ctx)
		}
	}
}

func RetryMiddleware(policy RetryPolicy) JobMiddleware {
	return func(next JobFunc) JobFunc {
		return func(ctx context.Context) error {
			var err error
			for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
				err = next(ctx)
				if err == nil {
					return nil
				}
				if attempt < policy.MaxRetries {
					time.Sleep(policy.Backoff * time.Duration(attempt+1))
				}
			}
			return fmt.Errorf("job failed after %d retries: %w", policy.MaxRetries, err)
		}
	}
}

type jobNameKey struct{}

func WithJobName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, jobNameKey{}, name)
}

func getJobName(ctx context.Context) string {
	if name, ok := ctx.Value(jobNameKey{}).(string); ok {
		return name
	}
	return "unknown"
}
