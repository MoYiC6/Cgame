package inventory

import (
	"context"

	"backend/internal/platform/database"
)

type Repository interface {
	Ping(ctx context.Context) error
}

type noopRepository struct {
	dbtx database.DBTX
}

func NewRepository() Repository {
	return &noopRepository{}
}

func (r *noopRepository) Ping(ctx context.Context) error {
	_ = r.executor(ctx)
	return nil
}

func (r *noopRepository) executor(ctx context.Context) database.DBTX {
	return database.ExecutorFromContext(ctx, r.dbtx)
}
