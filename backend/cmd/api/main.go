package main

import (
	"log"
	"os"

	"backend/internal/bootstrap"
	"backend/internal/modules/inventory"
	"backend/internal/modules/notification"
	"backend/internal/modules/order"
	"backend/internal/modules/payment"
	"backend/internal/platform/config"
	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
)

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
	deps := bootstrap.Dependencies{
		Config:     cfg,
		Logger:     appLogger,
		Tracer:     observability.NewNoopTracer(),
		Propagator: observability.NewNoopPropagator(),
	}

	engine := bootstrap.NewAPIEngine(
		deps,
		order.NewHandler(order.NewService(order.NewRepository())),
		payment.NewHandler(payment.NewService(payment.NewRepository())),
		inventory.NewHandler(inventory.NewService(inventory.NewRepository())),
		notification.NewHandler(notification.NewService(notification.NewRepository())),
	)

	appLogger.Info("api starting", "addr", cfg.Server.Addr, "config", cfg.MaskedSummary())
	if err := engine.Run(cfg.Server.Addr); err != nil {
		appLogger.Error("api stopped", "error", err)
		os.Exit(1)
	}
}
