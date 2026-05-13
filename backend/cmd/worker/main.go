package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/bootstrap"
	"backend/internal/platform/config"
	"backend/internal/platform/logger"
)

type placeholderTask struct{}

func (placeholderTask) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (placeholderTask) Probe(ctx context.Context) error {
	return nil
}

func main() {
	cfg, err := config.LoadConfig(os.Getenv("APP_ENV"))
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	appLogger := logger.New(cfg.Log.Level, os.Stdout)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	worker := bootstrap.NewWorker(appLogger)
	app := bootstrap.NewApp(worker)
	worker.RegisterRunnable("placeholder", placeholderTask{})

	appLogger.Info("worker starting", logger.Any("config", cfg.MaskedSummary()))
	if err := worker.Run(ctx); err != nil {
		appLogger.Error("worker stopped", "error", err)
		os.Exit(1)
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("worker shutdown failed", "error", err)
		os.Exit(1)
	}
	appLogger.Info("worker stopped cleanly")
}
