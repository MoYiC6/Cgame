package bootstrap

import (
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
		if deps.DB != nil {
			if err := deps.DB.Ping(c.Request.Context()); err != nil {
				response.Fail(c, err)
				return
			}
		}
		response.Success(c, gin.H{"status": "ok", "dependencies": "skipped"})
	})

	api := engine.Group("/api/v1")
	for _, registrar := range registrars {
		registrar.RegisterRoutes(api)
	}

	return engine
}
