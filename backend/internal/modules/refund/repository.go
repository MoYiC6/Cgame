package refund

import (
	"context"
	"fmt"
	"time"

	"backend/internal/platform/database"
)

type Repository interface {
	CreateRefund(ctx context.Context, refund *Refund) error
	ListUserRefunds(ctx context.Context, userID int64, page, pageSize int) (*RefundPageResult, error)
	GetRefundByID(ctx context.Context, id int64) (*Refund, error)
	GetRefundByOrderID(ctx context.Context, orderID int64) (*Refund, error)
	CancelRefund(ctx context.Context, id int64) error

	ListAdminRefunds(ctx context.Context, query RefundQuery) (*AdminRefundPageResult, error)
	UpdateStatus(ctx context.Context, id int64, status, adminRemark string, processedBy int64) error
	GetStats(ctx context.Context) (*RefundStats, error)
}

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

func (r *repository) CreateRefund(ctx context.Context, refund *Refund) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	refundNo := fmt.Sprintf("RF%s%06d", time.Now().Format("20060102150405"), refund.UserID%1000000)
	return exec.QueryRowContext(ctx,
		`INSERT INTO refunds (refund_no, order_id, user_id, amount, reason, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id`,
		refundNo, refund.OrderID, refund.UserID, refund.Amount, refund.Reason, RefundStatusPending,
	).Scan(&refund.ID)
}

func (r *repository) ListUserRefunds(ctx context.Context, userID int64, page, pageSize int) (*RefundPageResult, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var total int64
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM refunds WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count user refunds: %w", err)
	}

	rows, err := exec.QueryContext(ctx,
		`SELECT id, refund_no, order_id, amount, reason, status, created_at, updated_at
		 FROM refunds WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, fmt.Errorf("list user refunds: %w", err)
	}
	defer rows.Close()

	records := []RefundVO{}
	for rows.Next() {
		var item Refund
		if err := rows.Scan(&item.ID, &item.RefundNo, &item.OrderID, &item.Amount, &item.Reason, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan refund: %w", err)
		}
		records = append(records, toRefundVO(item))
	}
	return &RefundPageResult{Total: total, PageNum: page, PageSize: pageSize, Records: records}, nil
}

func (r *repository) GetRefundByID(ctx context.Context, id int64) (*Refund, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var item Refund
	if err := exec.QueryRowContext(ctx,
		`SELECT id, refund_no, order_id, user_id, amount, reason, status, admin_remark, processed_by, processed_at, created_at, updated_at
		 FROM refunds WHERE id = $1`,
		id,
	).Scan(&item.ID, &item.RefundNo, &item.OrderID, &item.UserID, &item.Amount, &item.Reason, &item.Status, &item.AdminRemark, &item.ProcessedBy, &item.ProcessedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get refund by id: %w", err)
	}
	return &item, nil
}

func (r *repository) GetRefundByOrderID(ctx context.Context, orderID int64) (*Refund, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var item Refund
	if err := exec.QueryRowContext(ctx,
		`SELECT id, refund_no, order_id, user_id, amount, reason, status, admin_remark, processed_by, processed_at, created_at, updated_at
		 FROM refunds WHERE order_id = $1`,
		orderID,
	).Scan(&item.ID, &item.RefundNo, &item.OrderID, &item.UserID, &item.Amount, &item.Reason, &item.Status, &item.AdminRemark, &item.ProcessedBy, &item.ProcessedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get refund by order id: %w", err)
	}
	return &item, nil
}

func (r *repository) CancelRefund(ctx context.Context, id int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx, `UPDATE refunds SET status = $1, updated_at = NOW() WHERE id = $2`, RefundStatusCancelled, id)
	if err != nil {
		return fmt.Errorf("cancel refund: %w", err)
	}
	return nil
}

func (r *repository) ListAdminRefunds(ctx context.Context, query RefundQuery) (*AdminRefundPageResult, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where, args := buildAdminWhere(query)

	var total int64
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM refunds "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count admin refunds: %w", err)
	}

	queryArgs := append([]any{}, args...)
	limitIndex := len(queryArgs) + 1
	offsetIndex := len(queryArgs) + 2
	queryArgs = append(queryArgs, query.PageSize, (query.PageNum-1)*query.PageSize)

	rows, err := exec.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, refund_no, order_id, user_id, amount, reason, status, admin_remark, processed_by, processed_at, created_at, updated_at
		 FROM refunds %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, limitIndex, offsetIndex),
		queryArgs...,
	)
	if err != nil {
		return nil, fmt.Errorf("list admin refunds: %w", err)
	}
	defer rows.Close()

	records := []AdminRefundVO{}
	for rows.Next() {
		var item Refund
		if err := rows.Scan(&item.ID, &item.RefundNo, &item.OrderID, &item.UserID, &item.Amount, &item.Reason, &item.Status, &item.AdminRemark, &item.ProcessedBy, &item.ProcessedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan admin refund: %w", err)
		}
		records = append(records, toAdminRefundVO(item))
	}
	return &AdminRefundPageResult{Total: total, PageNum: query.PageNum, PageSize: query.PageSize, Records: records}, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id int64, status, adminRemark string, processedBy int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE refunds SET status = $1, admin_remark = $2, processed_by = $3, processed_at = NOW(), updated_at = NOW() WHERE id = $4`,
		status, adminRemark, processedBy, id,
	)
	if err != nil {
		return fmt.Errorf("update refund status: %w", err)
	}
	return nil
}

func (r *repository) GetStats(ctx context.Context) (*RefundStats, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	stats := &RefundStats{}
	if err := exec.QueryRowContext(ctx,
		`SELECT COUNT(*),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'approved'),
			COUNT(*) FILTER (WHERE status = 'rejected'),
			COUNT(*) FILTER (WHERE status = 'processed'),
			COALESCE(SUM(amount), 0)
		 FROM refunds`,
	).Scan(&stats.TotalRefunds, &stats.PendingRefunds, &stats.ApprovedRefunds, &stats.RejectedRefunds, &stats.ProcessedRefunds, &stats.TotalAmount); err != nil {
		return nil, fmt.Errorf("get refund stats: %w", err)
	}
	return stats, nil
}

func buildAdminWhere(query RefundQuery) (string, []any) {
	where := "WHERE 1=1"
	args := []any{}

	if query.Status != "" {
		args = append(args, query.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if query.OrderID != nil {
		args = append(args, *query.OrderID)
		where += fmt.Sprintf(" AND order_id = $%d", len(args))
	}
	if query.UserID != nil {
		args = append(args, *query.UserID)
		where += fmt.Sprintf(" AND user_id = $%d", len(args))
	}
	if query.RefundNo != "" {
		args = append(args, "%"+query.RefundNo+"%")
		where += fmt.Sprintf(" AND refund_no ILIKE $%d", len(args))
	}
	if query.CreateTimeStart != nil && *query.CreateTimeStart != "" {
		args = append(args, *query.CreateTimeStart)
		where += fmt.Sprintf(" AND created_at >= $%d", len(args))
	}
	if query.CreateTimeEnd != nil && *query.CreateTimeEnd != "" {
		args = append(args, *query.CreateTimeEnd)
		where += fmt.Sprintf(" AND created_at <= $%d", len(args))
	}
	return where, args
}
