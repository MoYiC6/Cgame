package withdrawal

import (
	"context"
	"fmt"
	"math"
)

const defaultTaxRate = 0.06

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Client operations

func (s *Service) GetIncomeStats(ctx context.Context, teacherID int64) (*IncomeStats, error) {
	if teacherID == 0 {
		return nil, fmt.Errorf("teacher id is required")
	}
	return s.repo.GetTeacherStats(ctx, teacherID)
}

func (s *Service) CalculateWithdrawal(ctx context.Context, req CalculateRequest) (*CalculateResult, error) {
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}
	taxAmount := math.Round(req.Amount*defaultTaxRate*100) / 100
	actualAmount := math.Round((req.Amount-taxAmount)*100) / 100
	return &CalculateResult{
		Amount:       req.Amount,
		TaxAmount:    taxAmount,
		ActualAmount: actualAmount,
		TaxRate:      defaultTaxRate,
	}, nil
}

func (s *Service) Apply(ctx context.Context, teacherID int64, req ApplyRequest) (*Withdrawal, error) {
	if teacherID == 0 {
		return nil, fmt.Errorf("teacher id is required")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// Validate at least one payment method is provided
	if req.BankAccount == "" && req.AlipayAccount == "" {
		return nil, fmt.Errorf("bank account or alipay account is required")
	}

	stats, err := s.repo.GetTeacherStats(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("get teacher stats: %w", err)
	}
	if req.Amount > stats.UnsettledIncome {
		return nil, fmt.Errorf("insufficient unsettled income")
	}

	calc, err := s.CalculateWithdrawal(ctx, CalculateRequest{Amount: req.Amount})
	if err != nil {
		return nil, err
	}

	withdrawal := &Withdrawal{
		TeacherID:     teacherID,
		Amount:        req.Amount,
		TaxAmount:     calc.TaxAmount,
		ActualAmount:  calc.ActualAmount,
		Status:        WithdrawalStatusPending,
		BankName:      req.BankName,
		BankAccount:   req.BankAccount,
		AccountName:   req.AccountName,
		AlipayAccount: req.AlipayAccount,
		Remark:        req.Remark,
	}
	if err := s.repo.CreateWithdrawal(ctx, withdrawal); err != nil {
		return nil, fmt.Errorf("create withdrawal: %w", err)
	}
	return withdrawal, nil
}

func (s *Service) Cancel(ctx context.Context, teacherID, id int64) error {
	if teacherID == 0 || id == 0 {
		return fmt.Errorf("teacher id and withdrawal id are required")
	}
	w, err := s.repo.GetWithdrawalByID(ctx, id)
	if err != nil {
		return fmt.Errorf("withdrawal not found")
	}
	if w.TeacherID != teacherID {
		return fmt.Errorf("withdrawal does not belong to teacher")
	}
	if w.Status != WithdrawalStatusPending {
		return fmt.Errorf("only pending withdrawals can be cancelled")
	}
	return s.repo.CancelWithdrawal(ctx, id)
}

func (s *Service) ListMine(ctx context.Context, teacherID int64, page, pageSize int) (*WithdrawalPageResult, error) {
	if teacherID == 0 {
		return nil, fmt.Errorf("teacher id is required")
	}
	return s.repo.ListTeacherWithdrawals(ctx, teacherID, normalizePage(page), normalizePageSize(pageSize))
}

func (s *Service) GetMine(ctx context.Context, teacherID, id int64) (*WithdrawalVO, error) {
	if teacherID == 0 || id == 0 {
		return nil, fmt.Errorf("teacher id and withdrawal id are required")
	}
	w, err := s.repo.GetWithdrawalByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("withdrawal not found")
	}
	if w.TeacherID != teacherID {
		return nil, fmt.Errorf("withdrawal does not belong to teacher")
	}
	vo := toWithdrawalVO(*w)
	return &vo, nil
}

// Admin operations

func (s *Service) ListAdmin(ctx context.Context, query WithdrawalQuery) (*AdminWithdrawalPageResult, error) {
	query.PageNum = normalizePage(query.PageNum)
	query.PageSize = normalizePageSize(query.PageSize)
	return s.repo.ListAdminWithdrawals(ctx, query)
}

func (s *Service) GetAdmin(ctx context.Context, id int64) (*AdminWithdrawalVO, error) {
	if id == 0 {
		return nil, fmt.Errorf("withdrawal id is required")
	}
	w, err := s.repo.GetWithdrawalByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("withdrawal not found")
	}
	vo := toAdminWithdrawalVO(*w)
	return &vo, nil
}

func (s *Service) Approve(ctx context.Context, adminUserID, id int64, adminRemark string) error {
	if adminUserID == 0 || id == 0 {
		return fmt.Errorf("admin user id and withdrawal id are required")
	}
	w, err := s.repo.GetWithdrawalByID(ctx, id)
	if err != nil {
		return fmt.Errorf("withdrawal not found")
	}
	if w.Status != WithdrawalStatusPending {
		return fmt.Errorf("only pending withdrawals can be approved")
	}
	return s.repo.UpdateStatus(ctx, id, WithdrawalStatusApproved, adminRemark, adminUserID)
}

func (s *Service) Reject(ctx context.Context, adminUserID, id int64, adminRemark string) error {
	if adminUserID == 0 || id == 0 {
		return fmt.Errorf("admin user id and withdrawal id are required")
	}
	w, err := s.repo.GetWithdrawalByID(ctx, id)
	if err != nil {
		return fmt.Errorf("withdrawal not found")
	}
	if w.Status != WithdrawalStatusPending {
		return fmt.Errorf("only pending withdrawals can be rejected")
	}
	return s.repo.UpdateStatus(ctx, id, WithdrawalStatusRejected, adminRemark, adminUserID)
}

func (s *Service) Pay(ctx context.Context, adminUserID, id int64) error {
	if adminUserID == 0 || id == 0 {
		return fmt.Errorf("admin user id and withdrawal id are required")
	}
	w, err := s.repo.GetWithdrawalByID(ctx, id)
	if err != nil {
		return fmt.Errorf("withdrawal not found")
	}
	if w.Status != WithdrawalStatusApproved {
		return fmt.Errorf("only approved withdrawals can be paid")
	}
	return s.repo.UpdateStatus(ctx, id, WithdrawalStatusPaid, "", adminUserID)
}

func (s *Service) Stats(ctx context.Context) (*WithdrawalStats, error) {
	return s.repo.GetStats(ctx)
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
