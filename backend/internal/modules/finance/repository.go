package finance

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend/internal/platform/database"
)

type Repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) *Repository {
	return &Repository{dbtx: dbtx}
}

func (r *Repository) GetFinanceStats(ctx context.Context) (*FinanceStats, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	stats := &FinanceStats{}

	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status = 'completed'`,
	).Scan(&stats.TotalRevenue); err != nil {
		return nil, fmt.Errorf("get total revenue: %w", err)
	}

	today := time.Now().Truncate(24 * time.Hour)
	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status = 'completed' AND created_at >= $1`,
		today,
	).Scan(&stats.TodayRevenue); err != nil {
		return nil, fmt.Errorf("get today revenue: %w", err)
	}

	monthStart := time.Now().Truncate(24 * time.Hour).AddDate(0, 0, -time.Now().Day()+1)
	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status = 'completed' AND created_at >= $1`,
		monthStart,
	).Scan(&stats.MonthRevenue); err != nil {
		return nil, fmt.Errorf("get month revenue: %w", err)
	}

	yearStart := time.Now().Truncate(24 * time.Hour).AddDate(0, -int(time.Now().Month())+1, 0)
	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status = 'completed' AND created_at >= $1`,
		yearStart,
	).Scan(&stats.YearRevenue); err != nil {
		return nil, fmt.Errorf("get year revenue: %w", err)
	}

	stats.TeacherCommission = stats.TotalRevenue * 0.3
	stats.PlatformRevenue = stats.TotalRevenue * 0.7

	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status = 'paid'`,
	).Scan(&stats.PendingSettlement); err != nil {
		return nil, fmt.Errorf("get pending settlement: %w", err)
	}

	rows, err := exec.QueryContext(ctx,
		`SELECT TO_CHAR(created_at, 'YYYY-MM') as month, COALESCE(SUM(total_amount), 0) as revenue
		 FROM orders WHERE status = 'completed' AND created_at >= NOW() - INTERVAL '12 months'
		 GROUP BY TO_CHAR(created_at, 'YYYY-MM')
		 ORDER BY month ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("get monthly trend: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var trend MonthlyTrend
		if err := rows.Scan(&trend.Month, &trend.Revenue); err != nil {
			return nil, fmt.Errorf("scan trend: %w", err)
		}
		stats.MonthlyTrend = append(stats.MonthlyTrend, trend)
	}

	return stats, nil
}

func (r *Repository) GetMyCommissions(ctx context.Context, operatorID int64, page, pageSize int) ([]*OperatorCommission, int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var total int
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM operator_commissions WHERE operator_id = $1`, operatorID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count commissions: %w", err)
	}

	rows, err := exec.QueryContext(ctx,
		`SELECT id, operator_id, order_id, amount, balance, status, remark, created_at, updated_at
		 FROM operator_commissions WHERE operator_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		operatorID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list commissions: %w", err)
	}
	defer rows.Close()

	commissions := []*OperatorCommission{}
	for rows.Next() {
		var c OperatorCommission
		if err := rows.Scan(&c.ID, &c.OperatorID, &c.OrderID, &c.Amount, &c.Balance, &c.Status, &c.Remark, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan commission: %w", err)
		}
		commissions = append(commissions, &c)
	}
	return commissions, total, nil
}

func (r *Repository) GetMyCommissionBalance(ctx context.Context, operatorID int64) (float64, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var balance float64
	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(balance, 0) FROM operator_commissions WHERE operator_id = $1 ORDER BY created_at DESC LIMIT 1`,
		operatorID,
	).Scan(&balance); err != nil {
		return 0, nil
	}
	return balance, nil
}

func (r *Repository) ListOperatorWithdrawals(ctx context.Context, operatorID int64, page, pageSize int) ([]*OperatorWithdrawal, int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var total int
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM operator_withdrawals WHERE operator_id = $1`, operatorID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count withdrawals: %w", err)
	}

	rows, err := exec.QueryContext(ctx,
		`SELECT id, operator_id, amount, status, admin_remark, processed_by, processed_at, created_at, updated_at
		 FROM operator_withdrawals WHERE operator_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		operatorID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list withdrawals: %w", err)
	}
	defer rows.Close()

	withdrawals := []*OperatorWithdrawal{}
	for rows.Next() {
		var w OperatorWithdrawal
		if err := rows.Scan(&w.ID, &w.OperatorID, &w.Amount, &w.Status, &w.AdminRemark, &w.ProcessedBy, &w.ProcessedAt, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, &w)
	}
	return withdrawals, total, nil
}

func (r *Repository) CreateOperatorWithdrawal(ctx context.Context, withdrawal *OperatorWithdrawal) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO operator_withdrawals (operator_id, amount, status, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`,
		withdrawal.OperatorID, withdrawal.Amount, withdrawal.Status,
	).Scan(&withdrawal.ID)
}

func (r *Repository) UpdateOperatorWithdrawalStatus(ctx context.Context, id int64, status, adminRemark string, processedBy int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE operator_withdrawals SET status = $1, admin_remark = $2, processed_by = $3, processed_at = NOW(), updated_at = NOW() WHERE id = $4`,
		status, adminRemark, processedBy, id,
	)
	if err != nil {
		return fmt.Errorf("update withdrawal status: %w", err)
	}
	return nil
}

func (r *Repository) ListBalanceDetails(ctx context.Context, userID int64, page, pageSize int) ([]*BalanceDetail, int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var total int
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM balance_details WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count balance details: %w", err)
	}

	rows, err := exec.QueryContext(ctx,
		`SELECT id, user_id, change_type, amount, balance, remark, related_id, related_type, created_at
		 FROM balance_details WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list balance details: %w", err)
	}
	defer rows.Close()

	details := []*BalanceDetail{}
	for rows.Next() {
		var d BalanceDetail
		if err := rows.Scan(&d.ID, &d.UserID, &d.ChangeType, &d.Amount, &d.Balance, &d.Remark, &d.RelatedID, &d.RelatedType, &d.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan balance detail: %w", err)
		}
		details = append(details, &d)
	}
	return details, total, nil
}

func (r *Repository) GetMonthlyReport(ctx context.Context, month string) (*MonthlyReport, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	report := &MonthlyReport{Month: month}
	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(total_amount), 0), COUNT(*) FROM orders WHERE status = 'completed' AND TO_CHAR(created_at, 'YYYY-MM') = $1`,
		month,
	).Scan(&report.TotalRevenue, &report.TotalOrders); err != nil {
		return nil, fmt.Errorf("get monthly report: %w", err)
	}
	report.Commission = report.TotalRevenue * 0.3
	return report, nil
}

func (r *Repository) ListTeacherWithdrawals(ctx context.Context, query *TeacherWithdrawalQuery) ([]*TeacherWithdrawal, int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where := "WHERE 1=1"
	args := []any{}
	argIdx := 1

	if query.TeacherID > 0 {
		where += fmt.Sprintf(" AND teacher_id = $%d", argIdx)
		args = append(args, query.TeacherID)
		argIdx++
	}
	if query.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, query.Status)
		argIdx++
	}
	if query.StartDate != "" {
		where += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, query.StartDate)
		argIdx++
	}
	if query.EndDate != "" {
		where += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, query.EndDate+" 23:59:59")
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM teacher_withdrawals " + where
	if err := exec.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count teacher withdrawals: %w", err)
	}

	page := normalizePage(query.Page)
	pageSize := normalizePageSize(query.PageSize)
	limitOffset := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := exec.QueryContext(ctx,
		"SELECT id, teacher_id, amount, status, admin_remark, processed_by, processed_at, created_at, updated_at FROM teacher_withdrawals "+where+" ORDER BY created_at DESC "+limitOffset,
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list teacher withdrawals: %w", err)
	}
	defer rows.Close()

	withdrawals := []*TeacherWithdrawal{}
	for rows.Next() {
		var w TeacherWithdrawal
		if err := rows.Scan(&w.ID, &w.TeacherID, &w.Amount, &w.Status, &w.AdminRemark, &w.ProcessedBy, &w.ProcessedAt, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan teacher withdrawal: %w", err)
		}
		withdrawals = append(withdrawals, &w)
	}
	return withdrawals, total, nil
}

func (r *Repository) GetTeacherWithdrawalByID(ctx context.Context, id int64) (*TeacherWithdrawal, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var w TeacherWithdrawal
	if err := exec.QueryRowContext(ctx,
		`SELECT id, teacher_id, amount, status, admin_remark, processed_by, processed_at, created_at, updated_at
		 FROM teacher_withdrawals WHERE id = $1`,
		id,
	).Scan(&w.ID, &w.TeacherID, &w.Amount, &w.Status, &w.AdminRemark, &w.ProcessedBy, &w.ProcessedAt, &w.CreatedAt, &w.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("teacher withdrawal not found")
		}
		return nil, fmt.Errorf("get teacher withdrawal: %w", err)
	}
	return &w, nil
}

func (r *Repository) UpdateTeacherWithdrawalStatus(ctx context.Context, id int64, status, adminRemark string, processedBy int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE teacher_withdrawals SET status = $1, admin_remark = $2, processed_by = $3, processed_at = NOW(), updated_at = NOW() WHERE id = $4`,
		status, adminRemark, processedBy, id,
	)
	if err != nil {
		return fmt.Errorf("update teacher withdrawal status: %w", err)
	}
	return nil
}

func (r *Repository) GetWithdrawalStats(ctx context.Context) (*WithdrawalStats, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	stats := &WithdrawalStats{}

	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0), COUNT(*) FROM teacher_withdrawals`,
	).Scan(&stats.TotalWithdrawals, &stats.TotalCount); err != nil {
		return nil, fmt.Errorf("get total withdrawals: %w", err)
	}

	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0), COUNT(*) FROM teacher_withdrawals WHERE status = 'pending'`,
	).Scan(&stats.PendingAmount, &stats.PendingCount); err != nil {
		return nil, fmt.Errorf("get pending withdrawals: %w", err)
	}

	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0), COUNT(*) FROM teacher_withdrawals WHERE status = 'approved'`,
	).Scan(&stats.ApprovedAmount, &stats.ApprovedCount); err != nil {
		return nil, fmt.Errorf("get approved withdrawals: %w", err)
	}

	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0), COUNT(*) FROM teacher_withdrawals WHERE status = 'paid'`,
	).Scan(&stats.PaidAmount, &stats.PaidCount); err != nil {
		return nil, fmt.Errorf("get paid withdrawals: %w", err)
	}

	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0), COUNT(*) FROM teacher_withdrawals WHERE status = 'rejected'`,
	).Scan(&stats.RejectedAmount, &stats.RejectedCount); err != nil {
		return nil, fmt.Errorf("get rejected withdrawals: %w", err)
	}

	return stats, nil
}

func (r *Repository) ListSettleableOrders(ctx context.Context, teacherID int64, page, pageSize int) ([]*SettleableOrder, int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where := "WHERE o.status = 'completed' AND o.teacher_settled = false"
	args := []any{}
	argIdx := 1

	if teacherID > 0 {
		where += fmt.Sprintf(" AND o.teacher_id = $%d", argIdx)
		args = append(args, teacherID)
		argIdx++
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM orders o ` + where
	if err := exec.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count settleable orders: %w", err)
	}

	limitOffset := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := exec.QueryContext(ctx,
		`SELECT o.id, o.order_no, o.teacher_id, u.nickname, o.total_amount, o.teacher_commission, o.total_amount - o.teacher_commission, o.created_at, o.completed_at
		 FROM orders o
		 LEFT JOIN users u ON o.teacher_id = u.id
		 `+where+` ORDER BY o.completed_at DESC `+limitOffset,
		args...,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list settleable orders: %w", err)
	}
	defer rows.Close()

	orders := []*SettleableOrder{}
	for rows.Next() {
		var o SettleableOrder
		var completedAt sql.NullTime
		if err := rows.Scan(&o.ID, &o.OrderNo, &o.TeacherID, &o.TeacherName, &o.Amount, &o.Commission, &o.NetAmount, &o.CreatedAt, &completedAt); err != nil {
			return nil, 0, fmt.Errorf("scan settleable order: %w", err)
		}
		if completedAt.Valid {
			o.CompletedAt = completedAt.Time
		}
		orders = append(orders, &o)
	}
	return orders, total, nil
}

func (r *Repository) RejectOrderSettlement(ctx context.Context, withdrawalID, orderID int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE orders SET teacher_settled = true, settlement_rejected = true, settlement_rejected_at = NOW(), withdrawal_id = $1 WHERE id = $2`,
		withdrawalID, orderID,
	)
	if err != nil {
		return fmt.Errorf("reject order settlement: %w", err)
	}
	return nil
}

func (r *Repository) GetSettleableOrderTotal(ctx context.Context, teacherID int64) (float64, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var total float64
	where := "WHERE status = 'completed' AND teacher_settled = false"
	args := []any{}
	if teacherID > 0 {
		where += " AND teacher_id = $1"
		args = append(args, teacherID)
	}
	if err := exec.QueryRowContext(ctx,
		"SELECT COALESCE(SUM(teacher_commission), 0) FROM orders "+where,
		args...,
	).Scan(&total); err != nil {
		return 0, fmt.Errorf("get settleable order total: %w", err)
	}
	return total, nil
}

func (r *Repository) MarkOrdersAsSettled(ctx context.Context, teacherID int64, withdrawalID int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where := "WHERE status = 'completed' AND teacher_settled = false AND teacher_id = $1"
	_, err := exec.ExecContext(ctx,
		"UPDATE orders SET teacher_settled = true, settled_at = NOW(), withdrawal_id = $2 "+where,
		teacherID, withdrawalID,
	)
	if err != nil {
		return fmt.Errorf("mark orders as settled: %w", err)
	}
	return nil
}

func (r *Repository) CreateTeacherWithdrawal(ctx context.Context, withdrawal *TeacherWithdrawal) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO teacher_withdrawals (teacher_id, amount, status, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`,
		withdrawal.TeacherID, withdrawal.Amount, withdrawal.Status,
	).Scan(&withdrawal.ID)
}
