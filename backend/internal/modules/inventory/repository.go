package inventory

import "context"

type Repository interface {
	Ping(ctx context.Context) error
}

type noopRepository struct{}

func NewRepository() Repository {
	return noopRepository{}
}

func (noopRepository) Ping(ctx context.Context) error {
	return nil
}
