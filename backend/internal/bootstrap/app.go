package bootstrap

import (
	"log/slog"

	"backend/internal/platform/config"
	"backend/internal/platform/observability"
)

type Dependencies struct {
	Config     config.Config
	Logger     *slog.Logger
	Tracer     observability.Tracer
	Propagator observability.Propagator
}

type RouteRegistrar interface {
	RegisterRoutes(group any)
}
