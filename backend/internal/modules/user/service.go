package user

import (
	"context"
	"fmt"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetUser(ctx context.Context, userID int64) (*User, error) {
	return s.repo.GetByID(ctx, userID)
}

func (s *Service) CreateBalanceLog(ctx context.Context, log *UserBalanceLog) error {
	if log.UserID == 0 || log.ChangeType == "" {
		return fmt.Errorf("user_id and change_type are required")
	}
	return s.repo.CreateBalanceLog(ctx, log)
}

func (s *Service) GetBalanceLogs(ctx context.Context, userID int64, page, pageSize int) ([]*UserBalanceLog, int, error) {
	return s.repo.GetUserBalanceLogs(ctx, userID, page, pageSize)
}

func (s *Service) GetUserLevel(ctx context.Context, userID int64) (*UserLevel, error) {
	return s.repo.GetUserLevel(ctx, userID)
}

func (s *Service) CreateUserLevelLog(ctx context.Context, log *UserLevelLog) error {
	if log.UserID == 0 || log.NewLevelID == nil {
		return fmt.Errorf("user_id and new_level_id are required")
	}
	return s.repo.CreateUserLevelLog(ctx, log)
}

func (s *Service) CreatePurchaseRecord(ctx context.Context, record *UserPurchaseRecord) error {
	if record.UserID == 0 || record.GoodsID == nil {
		return fmt.Errorf("user_id and goods_id are required")
	}
	return s.repo.CreatePurchaseRecord(ctx, record)
}

func (s *Service) GetPurchaseCount(ctx context.Context, userID, goodsID int64) (int, error) {
	return s.repo.GetUserPurchaseCount(ctx, userID, goodsID)
}
