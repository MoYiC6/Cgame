package withdrawal

import (
	"context"
	"fmt"
	"time"

	"backend/internal/platform/database"
)

type Repository interface {
	CreateWithdrawal(ctx context.Context, withdrawal *Withdrawal) error
	GetWithdrawalByID(ctx context.Context, id int64) (*Withdrawal, error)
	ListTeacherWithdrawals(ctx context.Context, teacherID int64, page, pageSize int) (*WithdrawalPageResult, error)
	CancelWithdrawal(ctx context.Context, id int64) error

	ListAdminWithdrawals(ctx context.Context, query WithdrawalQuery) (*AdminWithdrawalPageResult, error)
	UpdateStatus(ctx context.Context, id int64, status, adminRemark string, processedBy int64) error
	GetStats(ctx context.Context) (*WithdrawalStats, error)
	GetTeacherStats(ctx context.Context, teacherID int64) (*IncomeStats, error)
}

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

func (r *repository) CreateWithdrawal(ctx context.Context, withdrawal *Withdrawal) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	withdrawalNo := fmt.Sprintf("WD%s%06d", time.Now().Format("20060102150405"), withdrawal.TeacherID%1000000)
	return exec.QueryRowContext(ctx,
		`INSERT INTO withdrawals (withdrawal_no, teacher_id, amount, tax_amount, actual_amount, status, bank_name, bank_account, account_name, alipay_account, remark, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW()) RETURNING id`,
		withdrawalNo, withdrawal.TeacherID, withdrawal.Amount, withdrawal.TaxAmount, withdrawal.ActualAmount, WithdrawalStatusPending,
		withdrawal.BankName, withdrawal.BankAccount, withdrawal.AccountName, withdrawal.AlipayAccount, withdrawal.Remark,
	).Scan(&withdrawal.ID)
}

func (r *repository) GetWithdrawalByID(ctx context.Context, id int64) (*Withdrawal, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var w Withdrawal
	if err := exec.QueryRowContext(ctx,
		`SELECT id, withdrawal_no, teacher_id, amount, tax_amount, actual_amount, status, bank_name, bank_account, account_name, alipay_account, remark, admin_remark, processed_by, processed_at, created_at, updated_at
		 FROM withdrawals WHERE id = $1`,
		id,
	).Scan(&w.ID, &w.WithdrawalNo, &w.TeacherID, &w.Amount, &w.TaxAmount, &w.ActualAmount, &w.Status, &w.BankName, &w.BankAccount, &w.AccountName, &w.AlipayAccount, &w.Remark, &w.AdminRemark, &w.ProcessedBy, &w.ProcessedAt, &w.CreatedAt, &w.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get withdrawal by id: %w", err)
	}
	return &w, nil
}

func (r *repository) ListTeacherWithdrawals(ctx context.Context, teacherID int64, page, pageSize int) (*WithdrawalPageResult, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var total int64
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM withdrawals WHERE teacher_id = $1`, teacherID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count teacher withdrawals: %w", err)
	}

	rows, err := exec.QueryContext(ctx,
		`SELECT id, withdrawal_no, amount, tax_amount, actual_amount, status, bank_name, bank_account, account_name, alipay_account, created_at
		 FROM withdrawals WHERE teacher_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		teacherID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, fmt.Errorf("list teacher withdrawals: %w", err)
	}
	defer rows.Close()

	records := []WithdrawalVO{}
	for rows.Next() {
		var w Withdrawal
		if err := rows.Scan(&w.ID, &w.WithdrawalNo, &w.Amount, &w.TaxAmount, &w.ActualAmount, &w.Status, &w.BankName, &w.BankAccount, &w.AccountName, &w.AlipayAccount, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan withdrawal: %w", err)
		}
		records = append(records, toWithdrawalVO(w))
	}
	return &WithdrawalPageResult{Total: total, PageNum: page, PageSize: pageSize, Records: records}, nil
}

func (r *repository) CancelWithdrawal(ctx context.Context, id int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx, `UPDATE withdrawals SET status = $1, updated_at = NOW() WHERE id = $2`, WithdrawalStatusCancelled, id)
	if err != nil {
		return fmt.Errorf("cancel withdrawal: %w", err)
	}
	return nil
}

func (r *repository) ListAdminWithdrawals(ctx context.Context, query WithdrawalQuery) (*AdminWithdrawalPageResult, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where, args := buildAdminWhere(query)

	var total int64
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM withdrawals "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count admin withdrawals: %w", err)
	}

	queryArgs := append([]any{}, args...)
	limitIndex := len(queryArgs) + 1
	offsetIndex := len(queryArgs) + 2
	queryArgs = append(queryArgs, query.PageSize, (query.PageNum-1)*query.PageSize)

	rows, err := exec.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, withdrawal_no, teacher_id, amount, tax_amount, actual_amount, status, bank_name, bank_account, account_name, alipay_account, remark, admin_remark, processed_by, processed_at, created_at, updated_at
		 FROM withdrawals %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, limitIndex, offsetIndex),
		queryArgs...,
	)
	if err != nil {
		return nil, fmt.Errorf("list admin withdrawals: %w", err)
	}
	defer rows.Close()

	records := []AdminWithdrawalVO{}
	for rows.Next() {
		var w Withdrawal
		if err := rows.Scan(&w.ID, &w.WithdrawalNo, &w.TeacherID, &w.Amount, &w.TaxAmount, &w.ActualAmount, &w.Status, &w.BankName, &w.BankAccount, &w.AccountName, &w.AlipayAccount, &w.Remark, &w.AdminRemark, &w.ProcessedBy, &w.ProcessedAt, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan admin withdrawal: %w", err)
		}
		records = append(records, toAdminWithdrawalVO(w))
	}
	return &AdminWithdrawalPageResult{Total: total, PageNum: query.PageNum, PageSize: query.PageSize, Records: records}, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id int64, status, adminRemark string, processedBy int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE withdrawals SET status = $1, admin_remark = $2, processed_by = $3, processed_at = NOW(), updated_at = NOW() WHERE id = $4`,
		status, adminRemark, processedBy, id,
	)
	if err != nil {
		return fmt.Errorf("update withdrawal status: %w", err)
	}
	return nil
}

func (r *repository) GetStats(ctx context.Context) (*WithdrawalStats, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	stats := &WithdrawalStats{}
	if err := exec.QueryRowContext(ctx,
		`SELECT COUNT(*),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'approved'),
			COUNT(*) FILTER (WHERE status = 'paid'),
			COALESCE(SUM(actual_amount), 0)
		 FROM withdrawals`,
	).Scan(&stats.TotalWithdrawals, &stats.PendingWithdrawals, &stats.ApprovedWithdrawals, &stats.PaidWithdrawals, &stats.TotalAmount); err != nil {
		return nil, fmt.Errorf("get withdrawal stats: %w", err)
	}
	return stats, nil
}

func (r *repository) GetTeacherStats(ctx context.Context, teacherID int64) (*IncomeStats, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	stats := &IncomeStats{}
	if err := exec.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0),
			COALESCE(SUM(CASE WHEN status = 'paid' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'pending' THEN amount ELSE 0 END), 0)
		 FROM withdrawals WHERE teacher_id = $1`,
		teacherID,
	).Scan(&stats.TotalIncome, &stats.WithdrawnAmount, &stats.PendingWithdrawal); err != nil {
		return nil, fmt.Errorf("get teacher stats: %w", err)
	}
	stats.SettledIncome = stats.WithdrawnAmount
	stats.UnsettledIncome = stats.TotalIncome - stats.WithdrawnAmount - stats.PendingWithdrawal
	if stats.UnsettledIncome < 0 {
		stats.UnsettledIncome = 0
	}
	return stats, nil
}

func buildAdminWhere(query WithdrawalQuery) (string, []any) {
	where := "WHERE 1=1"
	args := []any{}

	if query.Status != "" {
		args = append(args, query.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if query.TeacherID != nil {
		args = append(args, *query.TeacherID)
		where += fmt.Sprintf(" AND teacher_id = $%d", len(args))
	}
	if query.WithdrawalNo != "" {
		args = append(args, "%"+query.WithdrawalNo+"%")
		where += fmt.Sprintf(" AND withdrawal_no ILIKE $%d", len(args))
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
