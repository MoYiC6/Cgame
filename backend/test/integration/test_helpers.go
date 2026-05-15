package integration

import (
	"io"

	"backend/internal/bootstrap"
	"backend/internal/modules/inventory"
	"backend/internal/modules/notification"
	"backend/internal/modules/order"
	"backend/internal/modules/payment"
	"backend/internal/platform/config"
	"backend/internal/platform/database"
	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
	"github.com/gin-gonic/gin"
)

func newIntegrationEngine(db database.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	deps := bootstrap.Dependencies{
		Config: config.Config{
			App:    config.AppConfig{Name: "backend-test", Env: "test"},
			Server: config.ServerConfig{Addr: ":18080"},
			Log:    config.LogConfig{Level: "debug"},
		},
		Logger:     logger.New("debug", io.Discard),
		Tracer:     observability.NewNoopTracer(),
		Propagator: observability.NewNoopPropagator(),
		DB:         db,
	}

	return bootstrap.NewAPIEngine(
		deps,
		order.NewHandler(order.NewService(order.NewRepository(), database.NoopTxManager{})),
		payment.NewHandler(payment.NewService(payment.NewRepository(), database.NoopTxManager{})),
		inventory.NewHandler(inventory.NewService(inventory.NewRepository(), database.NoopTxManager{})),
		notification.NewHandler(notification.NewService(notification.NewRepository(), database.NoopTxManager{})),
	)
}
