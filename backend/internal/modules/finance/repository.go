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
