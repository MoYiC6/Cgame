package order

import (
	"context"

	"backend/internal/platform/database"
)

type Service interface {
	Ping(ctx context.Context) (PingResponse, error)
}

type service struct {
	repo      Repository
	txManager database.TxManager
}

func NewService(repo Repository, txManager database.TxManager) Service {
	if txManager == nil {
		txManager = database.NoopTxManager{}
	}
	return &service{repo: repo, txManager: txManager}
}

func (s *service) Ping(ctx context.Context) (PingResponse, error) {
	if err := s.repo.Ping(ctx); err != nil {
		return PingResponse{}, err
	}
	return PingResponse{Module: "order"}, nil
}
