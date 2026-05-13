package notification

import "context"

type Service interface {
	Ping(ctx context.Context) (PingResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Ping(ctx context.Context) (PingResponse, error) {
	if err := s.repo.Ping(ctx); err != nil {
		return PingResponse{}, err
	}
	return PingResponse{Module: "notification"}, nil
}
