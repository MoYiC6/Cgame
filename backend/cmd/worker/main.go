package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	configPath := os.Getenv("APP_CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.local.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	appLogger := logger.New(cfg.Log.Level, os.Stdout)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	worker := bootstrap.NewWorker(appLogger)
	worker.RegisterRunnable("placeholder", placeholderTask{})

	appLogger.Info("worker starting", "config", cfg.MaskedSummary())
	if err := worker.Run(ctx); err != nil {
		appLogger.Error("worker stopped", "error", err)
		os.Exit(1)
	}
	appLogger.Info("worker stopped cleanly")
}
