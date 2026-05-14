package bootstrap

import (
	"context"
	"errors"
	"net/http"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/database"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type HTTPRouteRegistrar interface {
	RegisterRoutes(group *gin.RouterGroup)
}

func NewAPIEngine(deps Dependencies, registrars ...HTTPRouteRegistrar) *gin.Engine {
	engine := gin.New()
	engine.Use(RequestIDMiddleware(), TraceContextMiddleware(deps.Propagator), RecoveryMiddleware(deps.Logger))

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
