package payment

import (
	"context"
	"fmt"
	"time"

	"backend/internal/platform/database"
)

type Service interface {
	Ping(ctx context.Context) (PingResponse, error)
	CreatePayment(ctx context.Context, userID int64, orderNo string, amount float64, payMethod string) (*PaymentRecord, error)
	ConfirmPayment(ctx context.Context, paymentNo string) error
	GetPayment(ctx context.Context, paymentNo string) (*PaymentRecord, error)
	ListPayments(ctx context.Context, userID int64, page, pageSize int) ([]*PaymentRecord, int, error)
	ListAdminPayments(ctx context.Context, page, pageSize int) ([]*PaymentRecord, int, error)
	GetPaymentStats(ctx context.Context) (*PaymentStats, error)
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
	return PingResponse{Module: "payment"}, nil
}

func (s *service) CreatePayment(ctx context.Context, userID int64, orderNo string, amount float64, payMethod string) (*PaymentRecord, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user_id is required")
	}
	if orderNo == "" {
		return nil, fmt.Errorf("order_no is required")
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be > 0")
	}

	paymentNo := fmt.Sprintf("PAY%d%d", time.Now().UnixNano(), userID)
	payment := &PaymentRecord{
		PaymentNo: paymentNo,
		OrderNo:   orderNo,
		UserID:    userID,
		Amount:    amount,
		Status:    "pending",
		PayMethod: payMethod,
	}

	if err := s.repo.CreatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}

	return payment, nil
}

func (s *service) ConfirmPayment(ctx context.Context, paymentNo string) error {
	if paymentNo == "" {
		return fmt.Errorf("payment_no is required")
	}

	payment, err := s.repo.GetPaymentByNo(ctx, paymentNo)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}
	if payment.Status != "pending" {
		return fmt.Errorf("payment cannot be confirmed")
	}

	now := time.Now()
	if err := s.repo.UpdatePaymentStatus(ctx, paymentNo, "paid"); err != nil {
		return fmt.Errorf("confirm payment: %w", err)
	}

	_ = now
	return nil
}

func (s *service) GetPayment(ctx context.Context, paymentNo string) (*PaymentRecord, error) {
	payment, err := s.repo.GetPaymentByNo(ctx, paymentNo)
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}
	return payment, nil
}

func (s *service) ListPayments(ctx context.Context, userID int64, page, pageSize int) ([]*PaymentRecord, int, error) {
	return s.repo.ListPayments(ctx, userID, page, pageSize)
}

func (s *service) ListAdminPayments(ctx context.Context, page, pageSize int) ([]*PaymentRecord, int, error) {
	return s.repo.ListAdminPayments(ctx, page, pageSize)
}

func (s *service) GetPaymentStats(ctx context.Context) (*PaymentStats, error) {
	return s.repo.GetPaymentStats(ctx)
}
