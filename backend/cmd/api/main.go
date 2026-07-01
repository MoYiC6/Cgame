package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/bootstrap"
	"backend/internal/modules/auth"
	"backend/internal/modules/inventory"
	"backend/internal/modules/notification"
	"backend/internal/modules/order"
	"backend/internal/modules/payment"
	"backend/internal/modules/user"
	"backend/internal/platform/config"
	"backend/internal/platform/database"
	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
	"backend/internal/platform/security"
)

func main() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	appLogger := logger.New(cfg.Log)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	provider, err := observability.InitProvider(ctx, cfg.Observability)
	if err != nil {
		if provider == nil {
			appLogger.Error("init observability provider failed", "error", err)
			os.Exit(1)
		}
		appLogger.Info("init observability provider degraded", "degraded", true, "error", err)
	}

	dbPool, err := database.NewPgxPool(ctx, cfg.DB)
	if err != nil {
		appLogger.Error("init db pool failed", "error", err)
		os.Exit(1)
	}

	sqlDB, err := dbPool.SQLDB()
	if err != nil {
		appLogger.Error("init sql db from pool failed", "error", err)
		os.Exit(1)
	}
	txManager := database.NewSQLTxManager(sqlDB)

	deps := bootstrap.Dependencies{
		Config:     *cfg,
		Logger:     appLogger,
		Tracer:     provider.Tracer(),
		Propagator: provider.Propagator(),
		DB:         dbPool,
	}

	passwordHasher := security.NewArgon2idHasher(
		cfg.Auth.Password.Argon2MemoryKiB,
		cfg.Auth.Password.Argon2Iterations,
		cfg.Auth.Password.Argon2Parallelism,
		os.Getenv("PASSWORD_PEPPER"),
	)
	tokenManager, err := newTokenManager(cfg)
	if err != nil {
		appLogger.Error("init token manager failed", "error", err)
		os.Exit(1)
	}
	authHandler := auth.NewHandler(
		auth.NewService(
			user.NewRepository(sqlDB),
			auth.NewRepository(sqlDB),
			txManager,
			passwordHasher,
			tokenManager,
			security.CryptoRandomTokenGenerator{},
			auth.ServiceConfig{RefreshTokenTTL: cfg.Auth.RefreshTokenTTL, RefreshCookieName: cfg.Auth.Cookie.Name, MaxFailedAttempts: cfg.Auth.Login.MaxFailedAttempts, FailedWindow: cfg.Auth.Login.FailedWindow, LockDuration: cfg.Auth.Login.LockDuration},
		),
		auth.NewHandlerConfigFromAuth(cfg.Auth),
		auth.AuthMiddleware(tokenManager),
	)

	engine := bootstrap.NewAPIEngine(
		deps,
		authHandler,
		order.NewHandler(order.NewService(order.NewRepository(), database.NoopTxManager{})),
		payment.NewHandler(payment.NewService(payment.NewRepository(), database.NoopTxManager{})),
		inventory.NewHandler(inventory.NewService(inventory.NewRepository(), database.NoopTxManager{})),
		notification.NewHandler(notification.NewService(notification.NewRepository(), database.NoopTxManager{})),
	)

	httpServer := bootstrap.NewHTTPServer(cfg.Server.Addr, engine)
	app := bootstrap.NewApp(httpServer, dbPool, sqlDB, provider)

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
}

func newTokenManager(cfg *config.Config) (security.TokenManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	if cfg.Auth.JWT.Algorithm != "HS256" {
		return nil, fmt.Errorf("unsupported jwt algorithm: %s", cfg.Auth.JWT.Algorithm)
	}
	return security.NewHMACTokenManager(security.HMACTokenConfig{
		Issuer:         cfg.Auth.Issuer,
		Audience:       cfg.Auth.Audience,
		KeyID:          cfg.Auth.JWT.KeyID,
		Secret:         []byte(os.Getenv("JWT_HMAC_SECRET")),
		AccessTokenTTL: cfg.Auth.AccessTokenTTL,
		ClockSkew:      30 * time.Second,
	}), nil
}
