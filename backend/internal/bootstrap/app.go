package bootstrap

import (
	"context"
	stderrors "errors"

	"backend/internal/platform/config"
	"backend/internal/platform/database"
	"backend/internal/platform/logger"
	"backend/internal/platform/observability"
)

type Dependencies struct {
	Config     config.Config
	Logger     logger.Logger
	Tracer     observability.Tracer
	Propagator observability.Propagator
	DB         database.DB
}

type RouteRegistrar interface {
	RegisterRoutes(group any)
}

type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

type App struct {
	shutdowners []Shutdowner
}

func NewApp(shutdowners ...Shutdowner) *App {
	return &App{shutdowners: shutdowners}
}

func (a *App) Shutdown(ctx context.Context) error {
	if a == nil {
		return nil
	}
	var errs []error
	for _, shutdowner := range a.shutdowners {
		if shutdowner == nil {
			continue
		}
		if err := shutdowner.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return stderrors.Join(errs...)
}
