package integration

import (
	"context"
	"database/sql"
	"io"
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
	"github.com/gin-gonic/gin"
)

type noopDBTX struct{}

func (noopDBTX) Exec(query string, args ...any) (sql.Result, error) { return nil, nil }
func (noopDBTX) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}
func (noopDBTX) Query(query string, args ...any) (*sql.Rows, error) { return nil, nil }
func (noopDBTX) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, nil
}
func (noopDBTX) QueryRow(query string, args ...any) *sql.Row { return &sql.Row{} }
func (noopDBTX) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return &sql.Row{}
}

func newIntegrationEngine(db database.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	repoDB := database.DBTX(noopDBTX{})
	if actual, ok := db.(database.DBTX); ok && actual != nil {
		repoDB = actual
	}
	deps := bootstrap.Dependencies{
		Config: config.Config{
			App:    config.AppConfig{Name: "backend-test", Env: "test"},
			Server: config.ServerConfig{Addr: ":18080"},
			Log:    config.LogConfig{Level: "debug"},
			Auth: config.AuthConfig{
				Issuer:          "backend",
				Audience:        "admin-api",
				AccessTokenTTL:  15 * time.Minute,
				RefreshTokenTTL: 24 * time.Hour,
				Login: config.AuthLoginConfig{
					MaxFailedAttempts: 5,
					FailedWindow:      15 * time.Minute,
					LockDuration:      30 * time.Minute,
				},
				Cookie: config.AuthCookieConfig{
					Name:     "refresh_token",
					Path:     "/api/v1/auth",
					HTTPOnly: true,
					SameSite: "lax",
				},
				Password: config.AuthPasswordConfig{
					MinLength:         12,
					MaxLength:         128,
					Argon2MemoryKiB:   19456,
					Argon2Iterations:  2,
					Argon2Parallelism: 1,
				},
				JWT: config.AuthJWTConfig{
					Algorithm: "HS256",
					KeyID:     "test-key",
				},
			},
		},
		Logger:     logger.NewText("debug", io.Discard),
		Tracer:     observability.NewNoopTracer(),
		Propagator: observability.NewNoopPropagator(),
		DB:         db,
	}

	passwordHasher := security.NewArgon2idHasher(
		deps.Config.Auth.Password.Argon2MemoryKiB,
		deps.Config.Auth.Password.Argon2Iterations,
		deps.Config.Auth.Password.Argon2Parallelism,
		"",
	)
	tokenManager := security.NewHMACTokenManager(security.HMACTokenConfig{
		Issuer:         deps.Config.Auth.Issuer,
		Audience:       deps.Config.Auth.Audience,
		KeyID:          deps.Config.Auth.JWT.KeyID,
		Secret:         []byte("01234567890123456789012345678901"),
		AccessTokenTTL: deps.Config.Auth.AccessTokenTTL,
		ClockSkew:      30 * time.Second,
	})
	authHandler := auth.NewHandler(
		auth.NewService(
			user.NewRepository(repoDB),
			auth.NewRepository(repoDB),
			database.NoopTxManager{},
			passwordHasher,
			tokenManager,
			security.CryptoRandomTokenGenerator{},
			auth.ServiceConfig{RefreshTokenTTL: deps.Config.Auth.RefreshTokenTTL, RefreshCookieName: deps.Config.Auth.Cookie.Name, MaxFailedAttempts: deps.Config.Auth.Login.MaxFailedAttempts, FailedWindow: deps.Config.Auth.Login.FailedWindow, LockDuration: deps.Config.Auth.Login.LockDuration},
		),
		auth.NewHandlerConfigFromAuth(deps.Config.Auth),
		auth.AuthMiddleware(tokenManager),
	)

	return bootstrap.NewAPIEngine(
		deps,
		authHandler,
		order.NewHandler(order.NewService(order.NewRepository(), database.NoopTxManager{})),
		payment.NewHandler(payment.NewService(payment.NewRepository(), database.NoopTxManager{})),
		inventory.NewHandler(inventory.NewService(inventory.NewRepository(), database.NoopTxManager{})),
		notification.NewHandler(notification.NewService(notification.NewRepository(repoDB), database.NoopTxManager{}), nil),
	)
}
