package user

import (
	"context"

	"backend/internal/platform/database"
)

type Repository interface {
	GetByID(ctx context.Context, userID int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type noopRepository struct {
	dbtx database.DBTX
}

func NewRepository() Repository {
	return &noopRepository{}
}

func (r *noopRepository) GetByID(ctx context.Context, userID int64) (*User, error) {
	_ = r.executor(ctx)
	return nil, nil
}

func (r *noopRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	_ = r.executor(ctx)
	return nil, nil
}

func (r *noopRepository) executor(ctx context.Context) database.DBTX {
	return database.ExecutorFromContext(ctx, r.dbtx)
}
