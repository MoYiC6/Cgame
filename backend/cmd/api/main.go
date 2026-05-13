package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/bootstrap"
	"backend/internal/modules/inventory"
	"backend/internal/modules/notification"
	"backend/internal/modules/order"
	"backend/internal/modules/payment"
	"backend/internal/platform/config"
	"backend/internal/platform/database"
	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
)

func main() {
	cfg, err := config.LoadConfig(os.Getenv("APP_ENV"))
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	appLogger := logger.New(cfg.Log.Level, os.Stdout)
	deps := bootstrap.Dependencies{
		Config:     *cfg,
		Logger:     appLogger,
		Tracer:     observability.NewNoopTracer(),
		Propagator: observability.NewNoopPropagator(),
		DB:         database.DummyDB{},
	}

	engine := bootstrap.NewAPIEngine(
		deps,
		order.NewHandler(order.NewService(order.NewRepository())),
		payment.NewHandler(payment.NewService(payment.NewRepository())),
		inventory.NewHandler(inventory.NewService(inventory.NewRepository())),
		notification.NewHandler(notification.NewService(notification.NewRepository())),
	)

	httpServer := bootstrap.NewHTTPServer(cfg.Server.Addr, engine)
	app := bootstrap.NewApp(httpServer)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := app.Shutdown(shutdownCtx); err != nil {
			appLogger.Error("api shutdown failed", "error", err)
		}
	}()

	appLogger.Info("api starting", logger.String("addr", cfg.Server.Addr), logger.Any("config", cfg.MaskedSummary()))
	if err := httpServer.Run(); err != nil {
		appLogger.Error("api stopped", "error", err)
		os.Exit(1)
	}
	appLogger.Info("api stopped cleanly")
}
