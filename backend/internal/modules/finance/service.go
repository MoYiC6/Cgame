package finance

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetFinanceStats(ctx context.Context) (*FinanceStats, error) {
	stats, err := s.repo.GetFinanceStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("get finance stats: %w", err)
	}
	return stats, nil
}
