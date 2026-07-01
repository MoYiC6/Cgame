package payment

import (
	"context"
	"fmt"

	"backend/internal/platform/database"
)

type Repository interface {
	Ping(ctx context.Context) error
	CreatePayment(ctx context.Context, payment *PaymentRecord) error
	GetPaymentByNo(ctx context.Context, paymentNo string) (*PaymentRecord, error)
	GetPaymentByOrderNo(ctx context.Context, orderNo string) (*PaymentRecord, error)
	ListPayments(ctx context.Context, userID int64, page, pageSize int) ([]*PaymentRecord, int, error)
	UpdatePaymentStatus(ctx context.Context, paymentNo string, status string) error
}

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

func (r *repository) Ping(ctx context.Context) error {
	return r.dbtx.QueryRowContext(ctx, "SELECT 1").Err()
}

func (r *repository) CreatePayment(ctx context.Context, payment *PaymentRecord) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO payment_records (payment_no, order_no, user_id, amount, status, pay_method)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		payment.PaymentNo, payment.OrderNo, payment.UserID, payment.Amount, payment.Status, payment.PayMethod,
	).Scan(&payment.ID)
}

func (r *repository) GetPaymentByNo(ctx context.Context, paymentNo string) (*PaymentRecord, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, payment_no, order_no, user_id, amount, status, pay_method, paid_at, refund_at, created_at, updated_at
		 FROM payment_records WHERE payment_no = $1`,
		paymentNo,
	)
	var payment PaymentRecord
	err := row.Scan(&payment.ID, &payment.PaymentNo, &payment.OrderNo, &payment.UserID, &payment.Amount, &payment.Status, &payment.PayMethod, &payment.PaidAt, &payment.RefundAt, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}
	return &payment, nil
}

func (r *repository) GetPaymentByOrderNo(ctx context.Context, orderNo string) (*PaymentRecord, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, payment_no, order_no, user_id, amount, status, pay_method, paid_at, refund_at, created_at, updated_at
		 FROM payment_records WHERE order_no = $1`,
		orderNo,
	)
	var payment PaymentRecord
	err := row.Scan(&payment.ID, &payment.PaymentNo, &payment.OrderNo, &payment.UserID, &payment.Amount, &payment.Status, &payment.PayMethod, &payment.PaidAt, &payment.RefundAt, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}
	return &payment, nil
}

func (r *repository) ListPayments(ctx context.Context, userID int64, page, pageSize int) ([]*PaymentRecord, int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM payment_records WHERE user_id = $1", userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count payments: %w", err)
	}

	rows, err := exec.QueryContext(ctx,
		`SELECT id, payment_no, order_no, user_id, amount, status, pay_method, paid_at, refund_at, created_at, updated_at
		 FROM payment_records WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()

	var payments []*PaymentRecord
	for rows.Next() {
		var p PaymentRecord
		if err := rows.Scan(&p.ID, &p.PaymentNo, &p.OrderNo, &p.UserID, &p.Amount, &p.Status, &p.PayMethod, &p.PaidAt, &p.RefundAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, &p)
	}
	return payments, total, nil
}

func (r *repository) UpdatePaymentStatus(ctx context.Context, paymentNo string, status string) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE payment_records SET status = $1, updated_at = NOW() WHERE payment_no = $2`,
		status, paymentNo,
	)
	if err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}
	return nil
}
