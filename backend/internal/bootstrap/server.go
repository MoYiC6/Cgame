package bootstrap

import (
	"context"
	"errors"
	"net/http"

	"backend/internal/platform/database"
	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPRouteRegistrar interface {
	RegisterRoutes(group *gin.RouterGroup)
}

func NewAPIEngine(deps Dependencies, registrars ...HTTPRouteRegistrar) *gin.Engine {
	engine := gin.New()
	engine.Use(
		RequestIDMiddleware(),
		TraceContextMiddleware(deps.Propagator),
		AccessLogMiddleware(deps.Logger),
		CORSMiddleware(deps.Config.CORS),
		SecurityHeadersMiddleware(deps.Config.SecurityHeaders),
		RateLimitMiddleware(deps.Config.RateLimit),
		RecoveryMiddleware(deps.Logger),
	)

	engine.GET("/healthz", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})
	engine.GET("/readyz", func(c *gin.Context) {
		if err := pingReadiness(c.Request.Context(), deps.DB); err != nil {
			response.Fail(c, apperrors.New("DEPENDENCY_UNAVAILABLE", "dependency unavailable", http.StatusServiceUnavailable, err))
			return
		}
		response.Success(c, gin.H{"status": "ok"})
	})
	if deps.Config.Metrics.Enabled {
		engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	api := engine.Group("/api/v1")
	for _, registrar := range registrars {
		registrar.RegisterRoutes(api)
	}

	return engine
}

func pingReadiness(ctx context.Context, db database.DB) error {
	if db == nil {
		return errors.New("db is required")
	}
	return db.Ping(ctx)
}
