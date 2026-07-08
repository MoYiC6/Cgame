package refund

import (
	"context"
	"fmt"
	"strings"

	"backend/internal/platform/database"
)

type OrderInfo struct {
	ID          int64
	UserID      int64
	Status      string
	TotalAmount float64
	PayAmount   float64
}

type OrderChecker interface {
	GetOrderByID(ctx context.Context, id int64) (*OrderInfo, error)
}

type Service struct {
	repo        Repository
	orderChecker OrderChecker
	txManager   database.TxManager
}

func NewService(repo Repository, txManager database.TxManager) *Service {
	if txManager == nil {
		txManager = database.NoopTxManager{}
	}
	return &Service{repo: repo, txManager: txManager}
}

func (s *Service) SetOrderChecker(checker OrderChecker) {
	s.orderChecker = checker
}

// Client operations

func (s *Service) Apply(ctx context.Context, userID int64, req ApplyRequest) (int64, error) {
	if userID == 0 {
		return 0, fmt.Errorf("user id is required")
	}
	if req.OrderID == 0 {
		return 0, fmt.Errorf("order id is required")
	}
	if req.Amount <= 0 {
		return 0, fmt.Errorf("refund amount must be greater than 0")
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return 0, fmt.Errorf("reason is required")
	}

	// Check if order exists and belongs to user
	if s.orderChecker != nil {
		order, err := s.orderChecker.GetOrderByID(ctx, req.OrderID)
		if err != nil {
			return 0, fmt.Errorf("order not found")
		}
		if order.UserID != userID {
			return 0, fmt.Errorf("order does not belong to user")
		}
		if req.Amount > order.PayAmount {
			return 0, fmt.Errorf("refund amount cannot exceed order pay amount")
		}
	}

	// Check if refund already exists for this order
	existing, err := s.repo.GetRefundByOrderID(ctx, req.OrderID)
	if err == nil && existing != nil {
		if existing.Status != RefundStatusCancelled && existing.Status != RefundStatusRejected {
			return 0, fmt.Errorf("refund already exists for this order")
		}
	}

	refund := &Refund{
		OrderID: req.OrderID,
		UserID:  userID,
		Amount:  req.Amount,
		Reason:  reason,
		Status:  RefundStatusPending,
	}
	if err := s.repo.CreateRefund(ctx, refund); err != nil {
		return 0, fmt.Errorf("create refund: %w", err)
	}
	return refund.ID, nil
}

func (s *Service) ListMine(ctx context.Context, userID int64, page, pageSize int) (*RefundPageResult, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is required")
	}
	return s.repo.ListUserRefunds(ctx, userID, normalizePage(page), normalizePageSize(pageSize))
}

func (s *Service) GetMine(ctx context.Context, userID, id int64) (*RefundVO, error) {
	if userID == 0 || id == 0 {
		return nil, fmt.Errorf("user id and refund id are required")
	}
	refund, err := s.repo.GetRefundByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("refund not found")
	}
	if refund.UserID != userID {
		return nil, fmt.Errorf("refund does not belong to user")
	}
	vo := toRefundVO(*refund)
	return &vo, nil
}

func (s *Service) Cancel(ctx context.Context, userID, id int64) error {
	if userID == 0 || id == 0 {
		return fmt.Errorf("user id and refund id are required")
	}
	refund, err := s.repo.GetRefundByID(ctx, id)
	if err != nil {
		return fmt.Errorf("refund not found")
	}
	if refund.UserID != userID {
		return fmt.Errorf("refund does not belong to user")
	}
	if refund.Status != RefundStatusPending {
		return fmt.Errorf("only pending refunds can be cancelled")
	}
	return s.repo.CancelRefund(ctx, id)
}

func (s *Service) CanApply(ctx context.Context, userID, orderID int64) (*CanApplyResult, error) {
	if orderID == 0 {
		return &CanApplyResult{CanApply: false, Reason: "order id is required"}, nil
	}
	if s.orderChecker != nil {
		order, err := s.orderChecker.GetOrderByID(ctx, orderID)
		if err != nil {
			return &CanApplyResult{CanApply: false, Reason: "order not found"}, nil
		}
		if order.UserID != userID {
			return &CanApplyResult{CanApply: false, Reason: "order does not belong to user"}, nil
		}
		if order.Status != "completed" && order.Status != "paid" {
			return &CanApplyResult{CanApply: false, Reason: "order status does not allow refund"}, nil
		}
	}
	existing, err := s.repo.GetRefundByOrderID(ctx, orderID)
	if err == nil && existing != nil {
		if existing.Status != RefundStatusCancelled && existing.Status != RefundStatusRejected {
			return &CanApplyResult{CanApply: false, Reason: "refund already exists for this order"}, nil
		}
	}
	return &CanApplyResult{CanApply: true}, nil
}

func (s *Service) GetByOrder(ctx context.Context, userID, orderID int64) (*RefundVO, error) {
	if userID == 0 || orderID == 0 {
		return nil, fmt.Errorf("user id and order id are required")
	}
	refund, err := s.repo.GetRefundByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("refund not found")
	}
	if refund.UserID != userID {
		return nil, fmt.Errorf("refund does not belong to user")
	}
	vo := toRefundVO(*refund)
	return &vo, nil
}

// Admin operations

func (s *Service) ListAdmin(ctx context.Context, query RefundQuery) (*AdminRefundPageResult, error) {
	query.PageNum = normalizePage(query.PageNum)
	query.PageSize = normalizePageSize(query.PageSize)
	return s.repo.ListAdminRefunds(ctx, query)
}

func (s *Service) GetAdmin(ctx context.Context, id int64) (*AdminRefundVO, error) {
	if id == 0 {
		return nil, fmt.Errorf("refund id is required")
	}
	refund, err := s.repo.GetRefundByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("refund not found")
	}
	vo := toAdminRefundVO(*refund)
	return &vo, nil
}

func (s *Service) Approve(ctx context.Context, adminUserID, id int64, adminRemark string) error {
	if adminUserID == 0 || id == 0 {
		return fmt.Errorf("admin user id and refund id are required")
	}
	refund, err := s.repo.GetRefundByID(ctx, id)
	if err != nil {
		return fmt.Errorf("refund not found")
	}
	if refund.Status != RefundStatusPending {
		return fmt.Errorf("only pending refunds can be approved")
	}
	return s.repo.UpdateStatus(ctx, id, RefundStatusApproved, strings.TrimSpace(adminRemark), adminUserID)
}

func (s *Service) Reject(ctx context.Context, adminUserID, id int64, adminRemark string) error {
	if adminUserID == 0 || id == 0 {
		return fmt.Errorf("admin user id and refund id are required")
	}
	refund, err := s.repo.GetRefundByID(ctx, id)
	if err != nil {
		return fmt.Errorf("refund not found")
	}
	if refund.Status != RefundStatusPending {
		return fmt.Errorf("only pending refunds can be rejected")
	}
	return s.repo.UpdateStatus(ctx, id, RefundStatusRejected, strings.TrimSpace(adminRemark), adminUserID)
}

func (s *Service) Process(ctx context.Context, adminUserID, id int64, adminRemark string) error {
	if adminUserID == 0 || id == 0 {
		return fmt.Errorf("admin user id and refund id are required")
	}
	refund, err := s.repo.GetRefundByID(ctx, id)
	if err != nil {
		return fmt.Errorf("refund not found")
	}
	if refund.Status != RefundStatusApproved {
		return fmt.Errorf("only approved refunds can be processed")
	}
	return s.repo.UpdateStatus(ctx, id, RefundStatusProcessed, strings.TrimSpace(adminRemark), adminUserID)
}

func (s *Service) Stats(ctx context.Context) (*RefundStats, error) {
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
