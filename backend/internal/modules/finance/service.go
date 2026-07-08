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

func (s *Service) ListTeacherWithdrawals(ctx context.Context, query *TeacherWithdrawalQuery) ([]*TeacherWithdrawal, int, error) {
	return s.repo.ListTeacherWithdrawals(ctx, query)
}

func (s *Service) GetTeacherWithdrawalByID(ctx context.Context, id int64) (*TeacherWithdrawal, error) {
	if id == 0 {
		return nil, fmt.Errorf("withdrawal id is required")
	}
	return s.repo.GetTeacherWithdrawalByID(ctx, id)
}

func (s *Service) ApproveTeacherWithdrawal(ctx context.Context, adminUserID, id int64, adminRemark string) error {
	if adminUserID == 0 || id == 0 {
		return fmt.Errorf("admin user id and withdrawal id are required")
	}
	return s.repo.UpdateTeacherWithdrawalStatus(ctx, id, "approved", adminRemark, adminUserID)
}

func (s *Service) RejectTeacherWithdrawal(ctx context.Context, adminUserID, id int64, adminRemark string) error {
	if adminUserID == 0 || id == 0 {
		return fmt.Errorf("admin user id and withdrawal id are required")
	}
	return s.repo.UpdateTeacherWithdrawalStatus(ctx, id, "rejected", adminRemark, adminUserID)
}

func (s *Service) PayTeacherWithdrawal(ctx context.Context, adminUserID, id int64) error {
	if adminUserID == 0 || id == 0 {
		return fmt.Errorf("admin user id and withdrawal id are required")
	}
	return s.repo.UpdateTeacherWithdrawalStatus(ctx, id, "paid", "", adminUserID)
}

func (s *Service) GetWithdrawalStats(ctx context.Context) (*WithdrawalStats, error) {
	return s.repo.GetWithdrawalStats(ctx)
}

func (s *Service) ListSettleableOrders(ctx context.Context, teacherID int64, page, pageSize int) ([]*SettleableOrder, int, error) {
	return s.repo.ListSettleableOrders(ctx, teacherID, normalizePage(page), normalizePageSize(pageSize))
}

func (s *Service) RejectOrderSettlement(ctx context.Context, adminUserID, withdrawalID, orderID int64) error {
	if adminUserID == 0 || withdrawalID == 0 || orderID == 0 {
		return fmt.Errorf("admin user id, withdrawal id and order id are required")
	}
	return s.repo.RejectOrderSettlement(ctx, withdrawalID, orderID)
}

func (s *Service) SettleOnBehalfPreview(ctx context.Context, teacherID int64) (float64, int, error) {
	if teacherID == 0 {
		return 0, 0, fmt.Errorf("teacher id is required")
	}
	total, err := s.repo.GetSettleableOrderTotal(ctx, teacherID)
	if err != nil {
		return 0, 0, fmt.Errorf("get settleable order total: %w", err)
	}
	orders, _, err := s.repo.ListSettleableOrders(ctx, teacherID, 1, 1000)
	if err != nil {
		return 0, 0, fmt.Errorf("list settleable orders: %w", err)
	}
	return total, len(orders), nil
}

func (s *Service) SettleOnBehalf(ctx context.Context, adminUserID, teacherID int64) (int64, error) {
	if adminUserID == 0 || teacherID == 0 {
		return 0, fmt.Errorf("admin user id and teacher id are required")
	}
	total, err := s.repo.GetSettleableOrderTotal(ctx, teacherID)
	if err != nil {
		return 0, fmt.Errorf("get settleable order total: %w", err)
	}
	if total <= 0 {
		return 0, fmt.Errorf("no settleable orders")
	}

	withdrawal := &TeacherWithdrawal{
		TeacherID: teacherID,
		Amount:    total,
		Status:    "pending",
	}
	if err := s.repo.CreateTeacherWithdrawal(ctx, withdrawal); err != nil {
		return 0, fmt.Errorf("create teacher withdrawal: %w", err)
	}
	if err := s.repo.MarkOrdersAsSettled(ctx, teacherID, withdrawal.ID); err != nil {
		return 0, fmt.Errorf("mark orders as settled: %w", err)
	}
	return withdrawal.ID, nil
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
