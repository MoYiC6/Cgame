package finance

import (
	"context"
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
