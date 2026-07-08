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

func (s *Service) GetMyCommissions(ctx context.Context, operatorID int64, page, pageSize int) ([]*OperatorCommission, int, error) {
	if operatorID == 0 {
		return nil, 0, fmt.Errorf("operator id is required")
	}
	return s.repo.GetMyCommissions(ctx, operatorID, normalizePage(page), normalizePageSize(pageSize))
}

func (s *Service) GetMyCommissionBalance(ctx context.Context, operatorID int64) (float64, error) {
	if operatorID == 0 {
		return 0, fmt.Errorf("operator id is required")
	}
	return s.repo.GetMyCommissionBalance(ctx, operatorID)
}

func (s *Service) ListMyWithdrawals(ctx context.Context, operatorID int64, page, pageSize int) ([]*OperatorWithdrawal, int, error) {
	if operatorID == 0 {
		return nil, 0, fmt.Errorf("operator id is required")
	}
	return s.repo.ListOperatorWithdrawals(ctx, operatorID, normalizePage(page), normalizePageSize(pageSize))
}

func (s *Service) ApplyWithdrawal(ctx context.Context, operatorID int64, amount float64) (int64, error) {
	if operatorID == 0 {
		return 0, fmt.Errorf("operator id is required")
	}
	if amount <= 0 {
		return 0, fmt.Errorf("amount must be greater than 0")
	}
	withdrawal := &OperatorWithdrawal{
		OperatorID: operatorID,
		Amount:     amount,
		Status:     "pending",
	}
	if err := s.repo.CreateOperatorWithdrawal(ctx, withdrawal); err != nil {
		return 0, fmt.Errorf("create withdrawal: %w", err)
	}
	return withdrawal.ID, nil
}

func (s *Service) ApproveWithdrawal(ctx context.Context, adminUserID, id int64, adminRemark string) error {
	if adminUserID == 0 || id == 0 {
		return fmt.Errorf("admin user id and withdrawal id are required")
	}
	return s.repo.UpdateOperatorWithdrawalStatus(ctx, id, "approved", adminRemark, adminUserID)
}

func (s *Service) RejectWithdrawal(ctx context.Context, adminUserID, id int64, adminRemark string) error {
	if adminUserID == 0 || id == 0 {
		return fmt.Errorf("admin user id and withdrawal id are required")
	}
	return s.repo.UpdateOperatorWithdrawalStatus(ctx, id, "rejected", adminRemark, adminUserID)
}

func (s *Service) PayWithdrawal(ctx context.Context, adminUserID, id int64) error {
	if adminUserID == 0 || id == 0 {
		return fmt.Errorf("admin user id and withdrawal id are required")
	}
	return s.repo.UpdateOperatorWithdrawalStatus(ctx, id, "paid", "", adminUserID)
}

func (s *Service) CancelWithdrawal(ctx context.Context, adminUserID, id int64) error {
	if adminUserID == 0 || id == 0 {
		return fmt.Errorf("admin user id and withdrawal id are required")
	}
	return s.repo.UpdateOperatorWithdrawalStatus(ctx, id, "cancelled", "", adminUserID)
}

func (s *Service) ListBalanceDetails(ctx context.Context, userID int64, page, pageSize int) ([]*BalanceDetail, int, error) {
	if userID == 0 {
		return nil, 0, fmt.Errorf("user id is required")
	}
	return s.repo.ListBalanceDetails(ctx, userID, normalizePage(page), normalizePageSize(pageSize))
}

func (s *Service) GetMonthlyReport(ctx context.Context, month string) (*MonthlyReport, error) {
	if month == "" {
		return nil, fmt.Errorf("month is required")
	}
	return s.repo.GetMonthlyReport(ctx, month)
}

func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func normalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return 10
	}
	if pageSize > 100 {
		return 100
	}
	return pageSize
}
