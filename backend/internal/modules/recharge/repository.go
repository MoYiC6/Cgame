package recharge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/internal/platform/database"
)

type Repository interface {
	CreateRechargeRecord(ctx context.Context, record *RechargeRecord) error
	GetRechargeByID(ctx context.Context, id int64) (*RechargeRecord, error)
	GetRechargeByNo(ctx context.Context, rechargeNo string) (*RechargeRecord, error)
	ListUserRecharges(ctx context.Context, userID int64, page, pageSize int) (*RechargeRecordPageResult, error)
	ListRecharges(ctx context.Context, query RechargeQuery) (*RechargeRecordPageResult, error)
	UpdateStatus(ctx context.Context, rechargeNo, status string, payChannel string, payTime, callbackTime *time.Time) error
	CancelRecharge(ctx context.Context, rechargeNo string) error
	GetStats(ctx context.Context) (*RechargeStats, error)
	GetRecentRecharges(ctx context.Context, userID int64, limit int) ([]RechargeRecordVO, error)

	CreateRebateRule(ctx context.Context, rule *RechargeRebateRule) error
	GetRebateRuleByID(ctx context.Context, id int64) (*RechargeRebateRule, error)
	ListRebateRules(ctx context.Context, page, pageSize int) (*RebateRulePageResult, error)
	UpdateRebateRule(ctx context.Context, id int64, updates map[string]any) error
	DeleteRebateRule(ctx context.Context, id int64) error
	ListEnabledRebateRules(ctx context.Context) ([]RechargeRebateRule, error)
}

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

func (r *repository) CreateRechargeRecord(ctx context.Context, record *RechargeRecord) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	rechargeNo := fmt.Sprintf("RC%s%06d", time.Now().Format("20060102150405"), record.UserID%1000000)
	return exec.QueryRowContext(ctx,
		`INSERT INTO recharge_records (recharge_no, user_id, amount, gift_amount, total_amount, pay_amount, status, pay_channel, remark, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) RETURNING id`,
		rechargeNo, record.UserID, record.Amount, record.GiftAmount, record.TotalAmount, record.PayAmount, record.Status, record.PayChannel, record.Remark,
	).Scan(&record.ID)
}

func (r *repository) GetRechargeByID(ctx context.Context, id int64) (*RechargeRecord, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var record RechargeRecord
	if err := exec.QueryRowContext(ctx,
		`SELECT id, recharge_no, user_id, amount, gift_amount, total_amount, pay_amount, status, pay_channel, pay_time, callback_time, remark, created_at, updated_at
		 FROM recharge_records WHERE id = $1`,
		id,
	).Scan(&record.ID, &record.RechargeNo, &record.UserID, &record.Amount, &record.GiftAmount, &record.TotalAmount, &record.PayAmount, &record.Status, &record.PayChannel, &record.PayTime, &record.CallbackTime, &record.Remark, &record.CreatedAt, &record.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get recharge by id: %w", err)
	}
	return &record, nil
}

func (r *repository) GetRechargeByNo(ctx context.Context, rechargeNo string) (*RechargeRecord, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var record RechargeRecord
	if err := exec.QueryRowContext(ctx,
		`SELECT id, recharge_no, user_id, amount, gift_amount, total_amount, pay_amount, status, pay_channel, pay_time, callback_time, remark, created_at, updated_at
		 FROM recharge_records WHERE recharge_no = $1`,
		rechargeNo,
	).Scan(&record.ID, &record.RechargeNo, &record.UserID, &record.Amount, &record.GiftAmount, &record.TotalAmount, &record.PayAmount, &record.Status, &record.PayChannel, &record.PayTime, &record.CallbackTime, &record.Remark, &record.CreatedAt, &record.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get recharge by no: %w", err)
	}
	return &record, nil
}

func (r *repository) ListUserRecharges(ctx context.Context, userID int64, page, pageSize int) (*RechargeRecordPageResult, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var total int64
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM recharge_records WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, fmt.Errorf("count user recharges: %w", err)
	}

	rows, err := exec.QueryContext(ctx,
		`SELECT id, recharge_no, amount, gift_amount, total_amount, status, pay_channel, pay_time, created_at
		 FROM recharge_records WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, fmt.Errorf("list user recharges: %w", err)
	}
	defer rows.Close()

	records := []RechargeRecordVO{}
	for rows.Next() {
		var item RechargeRecord
		if err := rows.Scan(&item.ID, &item.RechargeNo, &item.Amount, &item.GiftAmount, &item.TotalAmount, &item.Status, &item.PayChannel, &item.PayTime, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan recharge: %w", err)
		}
		records = append(records, toRechargeRecordVO(item))
	}
	return &RechargeRecordPageResult{Total: total, PageNum: page, PageSize: pageSize, Records: records}, nil
}

func (r *repository) ListRecharges(ctx context.Context, query RechargeQuery) (*RechargeRecordPageResult, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where, args := buildAdminWhere(query)

	var total int64
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM recharge_records "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count recharges: %w", err)
	}

	queryArgs := append([]any{}, args...)
	limitIndex := len(queryArgs) + 1
	offsetIndex := len(queryArgs) + 2
	queryArgs = append(queryArgs, query.PageSize, (query.PageNum-1)*query.PageSize)

	rows, err := exec.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, recharge_no, user_id, amount, gift_amount, total_amount, status, pay_channel, pay_time, created_at
		 FROM recharge_records %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, limitIndex, offsetIndex),
		queryArgs...,
	)
	if err != nil {
		return nil, fmt.Errorf("list recharges: %w", err)
	}
	defer rows.Close()

	records := []RechargeRecordVO{}
	for rows.Next() {
		var item RechargeRecord
		if err := rows.Scan(&item.ID, &item.RechargeNo, &item.UserID, &item.Amount, &item.GiftAmount, &item.TotalAmount, &item.Status, &item.PayChannel, &item.PayTime, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan recharge: %w", err)
		}
		records = append(records, toRechargeRecordVO(item))
	}
	return &RechargeRecordPageResult{Total: total, PageNum: query.PageNum, PageSize: query.PageSize, Records: records}, nil
}

func (r *repository) UpdateStatus(ctx context.Context, rechargeNo, status string, payChannel string, payTime, callbackTime *time.Time) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE recharge_records SET status = $1, pay_channel = $2, pay_time = $3, callback_time = $4, updated_at = NOW() WHERE recharge_no = $5`,
		status, payChannel, payTime, callbackTime, rechargeNo,
	)
	if err != nil {
		return fmt.Errorf("update recharge status: %w", err)
	}
	return nil
}

func (r *repository) CancelRecharge(ctx context.Context, rechargeNo string) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE recharge_records SET status = $1, updated_at = NOW() WHERE recharge_no = $2`,
		RechargeStatusCancelled, rechargeNo,
	)
	if err != nil {
		return fmt.Errorf("cancel recharge: %w", err)
	}
	return nil
}

func (r *repository) GetStats(ctx context.Context) (*RechargeStats, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	stats := &RechargeStats{}
	if err := exec.QueryRowContext(ctx,
		`SELECT COUNT(*),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'paid'),
			COALESCE(SUM(amount), 0),
			COALESCE(SUM(gift_amount), 0)
		 FROM recharge_records`,
	).Scan(&stats.TotalRecords, &stats.PendingRecords, &stats.PaidRecords, &stats.TotalAmount, &stats.TotalGiftAmount); err != nil {
		return nil, fmt.Errorf("get recharge stats: %w", err)
	}
	return stats, nil
}

func (r *repository) GetRecentRecharges(ctx context.Context, userID int64, limit int) ([]RechargeRecordVO, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	rows, err := exec.QueryContext(ctx,
		`SELECT id, recharge_no, amount, gift_amount, total_amount, status, pay_channel, pay_time, created_at
		 FROM recharge_records WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("get recent recharges: %w", err)
	}
	defer rows.Close()

	records := []RechargeRecordVO{}
	for rows.Next() {
		var item RechargeRecord
		if err := rows.Scan(&item.ID, &item.RechargeNo, &item.Amount, &item.GiftAmount, &item.TotalAmount, &item.Status, &item.PayChannel, &item.PayTime, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan recharge: %w", err)
		}
		records = append(records, toRechargeRecordVO(item))
	}
	return records, nil
}

// Rebate rule operations

func (r *repository) CreateRebateRule(ctx context.Context, rule *RechargeRebateRule) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO recharge_rebate_rules (name, min_amount, gift_rate, gift_amount, enabled, priority, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id`,
		rule.Name, rule.MinAmount, rule.GiftRate, rule.GiftAmount, rule.Enabled, rule.Priority,
	).Scan(&rule.ID)
}

func (r *repository) GetRebateRuleByID(ctx context.Context, id int64) (*RechargeRebateRule, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var rule RechargeRebateRule
	if err := exec.QueryRowContext(ctx,
		`SELECT id, name, min_amount, gift_rate, gift_amount, enabled, priority, created_at, updated_at
		 FROM recharge_rebate_rules WHERE id = $1`,
		id,
	).Scan(&rule.ID, &rule.Name, &rule.MinAmount, &rule.GiftRate, &rule.GiftAmount, &rule.Enabled, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get rebate rule by id: %w", err)
	}
	return &rule, nil
}

func (r *repository) ListRebateRules(ctx context.Context, page, pageSize int) (*RebateRulePageResult, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var total int64
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(*) FROM recharge_rebate_rules`).Scan(&total); err != nil {
		return nil, fmt.Errorf("count rebate rules: %w", err)
	}

	rows, err := exec.QueryContext(ctx,
		`SELECT id, name, min_amount, gift_rate, gift_amount, enabled, priority, created_at, updated_at
		 FROM recharge_rebate_rules ORDER BY priority DESC, created_at DESC LIMIT $1 OFFSET $2`,
		pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, fmt.Errorf("list rebate rules: %w", err)
	}
	defer rows.Close()

	records := []RechargeRebateRuleVO{}
	for rows.Next() {
		var rule RechargeRebateRule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.MinAmount, &rule.GiftRate, &rule.GiftAmount, &rule.Enabled, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan rebate rule: %w", err)
		}
		records = append(records, toRebateRuleVO(rule))
	}
	return &RebateRulePageResult{Total: total, PageNum: page, PageSize: pageSize, Records: records}, nil
}

func (r *repository) UpdateRebateRule(ctx context.Context, id int64, updates map[string]any) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	if len(updates) == 0 {
		return nil
	}
	sets := []string{}
	args := []any{}
	for col, val := range updates {
		args = append(args, val)
		sets = append(sets, fmt.Sprintf("%s = $%d", col, len(args)))
	}
	args = append(args, id)
	query := fmt.Sprintf("UPDATE recharge_rebate_rules SET %s, updated_at = NOW() WHERE id = $%d",
		strings.Join(sets, ", "), len(args))
	_, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update rebate rule: %w", err)
	}
	return nil
}

func (r *repository) DeleteRebateRule(ctx context.Context, id int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx, `DELETE FROM recharge_rebate_rules WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete rebate rule: %w", err)
	}
	return nil
}

func (r *repository) ListEnabledRebateRules(ctx context.Context) ([]RechargeRebateRule, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	rows, err := exec.QueryContext(ctx,
		`SELECT id, name, min_amount, gift_rate, gift_amount, enabled, priority, created_at, updated_at
		 FROM recharge_rebate_rules WHERE enabled = TRUE ORDER BY priority DESC, min_amount DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list enabled rebate rules: %w", err)
	}
	defer rows.Close()

	rules := []RechargeRebateRule{}
	for rows.Next() {
		var rule RechargeRebateRule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.MinAmount, &rule.GiftRate, &rule.GiftAmount, &rule.Enabled, &rule.Priority, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan rebate rule: %w", err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func buildAdminWhere(query RechargeQuery) (string, []any) {
	where := "WHERE 1=1"
	args := []any{}

	if query.Status != "" {
		args = append(args, query.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if query.UserID != nil {
		args = append(args, *query.UserID)
		where += fmt.Sprintf(" AND user_id = $%d", len(args))
	}
	if query.RechargeNo != "" {
		args = append(args, "%"+query.RechargeNo+"%")
		where += fmt.Sprintf(" AND recharge_no ILIKE $%d", len(args))
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

// RebateRulePageResult wraps paginated rebate rules.
type RebateRulePageResult struct {
	Total    int64               `json:"total"`
	PageNum  int                 `json:"pageNum"`
	PageSize int                 `json:"pageSize"`
	Records  []RechargeRebateRuleVO `json:"records"`
}
