package teacher

import (
	"context"
	"encoding/json"
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

func (r *Repository) CreateTeacher(ctx context.Context, t *Teacher) error {
	platformsJSON, _ := json.Marshal(t.Platforms)
	tagsJSON, _ := json.Marshal(t.Tags)
	goodsIDsJSON, _ := json.Marshal(t.GoodsIDs)
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO teachers (user_id, name, mobile, avatar, status, rating, order_count, deposit, balance, platforms, tags, goods_ids, auto_status_enabled)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) RETURNING id`,
		t.UserID, t.Name, t.Mobile, t.Avatar, t.Status, t.Rating, t.OrderCount, t.Deposit, t.Balance, string(platformsJSON), string(tagsJSON), string(goodsIDsJSON), t.AutoStatusEnabled,
	).Scan(&t.ID)
}

func (r *Repository) GetTeacherByUserID(ctx context.Context, userID int64) (*Teacher, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, user_id, name, mobile, avatar, status, rating, order_count, deposit, balance, platforms, tags, goods_ids, auto_status_enabled, created_at, updated_at
		 FROM teachers WHERE user_id = $1`,
		userID,
	)
	var t Teacher
	var platformsJSON, tagsJSON, goodsIDsJSON []byte
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.Mobile, &t.Avatar, &t.Status, &t.Rating, &t.OrderCount, &t.Deposit, &t.Balance, &platformsJSON, &tagsJSON, &goodsIDsJSON, &t.AutoStatusEnabled, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get teacher: %w", err)
	}
	if len(platformsJSON) > 0 {
		json.Unmarshal(platformsJSON, &t.Platforms)
	}
	if len(tagsJSON) > 0 {
		json.Unmarshal(tagsJSON, &t.Tags)
	}
	if len(goodsIDsJSON) > 0 {
		json.Unmarshal(goodsIDsJSON, &t.GoodsIDs)
	}
	return &t, nil
}

func (r *Repository) GetTeacherByID(ctx context.Context, id int64) (*Teacher, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, user_id, name, mobile, avatar, status, rating, order_count, deposit, balance, platforms, tags, goods_ids, auto_status_enabled, created_at, updated_at
		 FROM teachers WHERE id = $1`,
		id,
	)
	var t Teacher
	var platformsJSON, tagsJSON, goodsIDsJSON []byte
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.Mobile, &t.Avatar, &t.Status, &t.Rating, &t.OrderCount, &t.Deposit, &t.Balance, &platformsJSON, &tagsJSON, &goodsIDsJSON, &t.AutoStatusEnabled, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get teacher by id: %w", err)
	}
	if len(platformsJSON) > 0 {
		json.Unmarshal(platformsJSON, &t.Platforms)
	}
	if len(tagsJSON) > 0 {
		json.Unmarshal(tagsJSON, &t.Tags)
	}
	if len(goodsIDsJSON) > 0 {
		json.Unmarshal(goodsIDsJSON, &t.GoodsIDs)
	}
	return &t, nil
}

func (r *Repository) ListTeachers(ctx context.Context, page, pageSize int) ([]*Teacher, int, error) {
	countQuery := "SELECT COUNT(*) FROM teachers WHERE status = 1"
	var total int
	if err := r.dbtx.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, user_id, name, mobile, avatar, status, rating, order_count, deposit, balance, platforms, tags, goods_ids, auto_status_enabled, created_at, updated_at
		 FROM teachers WHERE status = 1 ORDER BY order_count DESC LIMIT $1 OFFSET $2`,
		pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var teachers []*Teacher
	for rows.Next() {
		var t Teacher
		var platformsJSON, tagsJSON, goodsIDsJSON []byte
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Mobile, &t.Avatar, &t.Status, &t.Rating, &t.OrderCount, &t.Deposit, &t.Balance, &platformsJSON, &tagsJSON, &goodsIDsJSON, &t.AutoStatusEnabled, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if len(platformsJSON) > 0 {
			json.Unmarshal(platformsJSON, &t.Platforms)
		}
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &t.Tags)
		}
		if len(goodsIDsJSON) > 0 {
			json.Unmarshal(goodsIDsJSON, &t.GoodsIDs)
		}
		teachers = append(teachers, &t)
	}
	return teachers, total, nil
}

func (r *Repository) UpdateTeacherStatus(ctx context.Context, teacherID int64, status int) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE teachers SET status = $2, updated_at = NOW() WHERE id = $1`,
		teacherID, status,
	)
	return err
}

func (r *Repository) UpdateTeacherHeartbeat(ctx context.Context, teacherID int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE teachers SET updated_at = NOW() WHERE id = $1`,
		teacherID,
	)
	return err
}

func (r *Repository) UpdateTeacherAutoStatus(ctx context.Context, teacherID int64, enabled bool, onlineTime, offlineTime *string) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE teachers SET auto_status_enabled = $2, updated_at = NOW() WHERE id = $1`,
		teacherID, enabled,
	)
	return err
}

func (r *Repository) CreateTeacherLevel(ctx context.Context, level *TeacherLevel) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO teacher_levels (name, min_orders, commission_rate, priority, status, description) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		level.Name, level.MinOrders, level.CommissionRate, level.Priority, level.Status, level.Description,
	).Scan(&level.ID)
}

func (r *Repository) GetTeacherLevels(ctx context.Context) ([]*TeacherLevel, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, name, min_orders, commission_rate, priority, status, description, created_at, updated_at FROM teacher_levels WHERE status = 1 ORDER BY priority ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []*TeacherLevel
	for rows.Next() {
		var l TeacherLevel
		if err := rows.Scan(&l.ID, &l.Name, &l.MinOrders, &l.CommissionRate, &l.Priority, &l.Status, &l.Description, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		levels = append(levels, &l)
	}
	return levels, nil
}

func (r *Repository) UpdateTeacherLevel(ctx context.Context, level *TeacherLevel) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE teacher_levels SET name = $2, min_orders = $3, commission_rate = $4, priority = $5, status = $6, description = $7, updated_at = NOW() WHERE id = $1`,
		level.ID, level.Name, level.MinOrders, level.CommissionRate, level.Priority, level.Status, level.Description,
	)
	return err
}

func (r *Repository) DeleteTeacherLevel(ctx context.Context, id int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		`DELETE FROM teacher_levels WHERE id = $1`,
		id,
	)
	return err
}

func (r *Repository) CreateTeacherStatusLog(ctx context.Context, log *TeacherStatusLog) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO teacher_status_logs (teacher_id, old_status, new_status, reason, operator_id) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		log.TeacherID, log.OldStatus, log.NewStatus, log.Reason, log.OperatorID,
	).Scan(&log.ID)
}

func (r *Repository) GetTeacherStatusLogs(ctx context.Context, teacherID int64, page, pageSize int) ([]*TeacherStatusLog, int, error) {
	var total int
	if err := r.dbtx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM teacher_status_logs WHERE teacher_id = $1`,
		teacherID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, teacher_id, old_status, new_status, reason, operator_id, created_at
		 FROM teacher_status_logs WHERE teacher_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		teacherID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*TeacherStatusLog
	for rows.Next() {
		var l TeacherStatusLog
		if err := rows.Scan(&l.ID, &l.TeacherID, &l.OldStatus, &l.NewStatus, &l.Reason, &l.OperatorID, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, &l)
	}
	return logs, total, nil
}

func (r *Repository) CreateTeacherBalanceLog(ctx context.Context, log *TeacherBalanceLog) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO teacher_balance_logs (teacher_id, change_type, amount, balance_after, related_id, related_no, description) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		log.TeacherID, log.ChangeType, log.Amount, log.BalanceAfter, log.RelatedID, log.RelatedNo, log.Description,
	).Scan(&log.ID)
}

// TeacherApplication CRUD

func (r *Repository) CreateTeacherApplication(ctx context.Context, app *TeacherApplication) (int64, error) {
	platformsJSON, _ := json.Marshal(app.Platforms)
	tagsJSON, _ := json.Marshal(app.Tags)
	err := r.dbtx.QueryRowContext(ctx,
		`INSERT INTO teacher_applications (user_id, name, mobile, avatar, platforms, tags, intro, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()) RETURNING id`,
		app.UserID, app.Name, app.Mobile, app.Avatar, string(platformsJSON), string(tagsJSON), app.Intro, app.Status,
	).Scan(&app.ID)
	return app.ID, err
}

func (r *Repository) GetTeacherApplicationByID(ctx context.Context, id int64) (*TeacherApplication, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, user_id, name, mobile, avatar, platforms, tags, intro, status, reason, operator_id, created_at, updated_at
		 FROM teacher_applications WHERE id = $1`,
		id,
	)
	var app TeacherApplication
	var platformsJSON, tagsJSON []byte
	err := row.Scan(&app.ID, &app.UserID, &app.Name, &app.Mobile, &app.Avatar, &platformsJSON, &tagsJSON, &app.Intro, &app.Status, &app.Reason, &app.OperatorID, &app.CreatedAt, &app.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get application by id: %w", err)
	}
	if len(platformsJSON) > 0 {
		json.Unmarshal(platformsJSON, &app.Platforms)
	}
	if len(tagsJSON) > 0 {
		json.Unmarshal(tagsJSON, &app.Tags)
	}
	return &app, nil
}

func (r *Repository) ListTeacherApplications(ctx context.Context, status *int, page, pageSize int) ([]*TeacherApplication, int, error) {
	countQuery := `SELECT COUNT(*) FROM teacher_applications`
	var total int
	var args []any
	if status != nil {
		countQuery += ` WHERE status = $1`
		args = append(args, *status)
	}
	if err := r.dbtx.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, user_id, name, mobile, avatar, platforms, tags, intro, status, reason, operator_id, created_at, updated_at
			  FROM teacher_applications`
	args = nil
	if status != nil {
		query += ` WHERE status = $1`
		args = append(args, *status)
	}
	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.dbtx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var apps []*TeacherApplication
	for rows.Next() {
		var app TeacherApplication
		var platformsJSON, tagsJSON []byte
		if err := rows.Scan(&app.ID, &app.UserID, &app.Name, &app.Mobile, &app.Avatar, &platformsJSON, &tagsJSON, &app.Intro, &app.Status, &app.Reason, &app.OperatorID, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if len(platformsJSON) > 0 {
			json.Unmarshal(platformsJSON, &app.Platforms)
		}
		if len(tagsJSON) > 0 {
			json.Unmarshal(tagsJSON, &app.Tags)
		}
		apps = append(apps, &app)
	}
	return apps, total, nil
}

func (r *Repository) UpdateApplicationStatus(ctx context.Context, id int64, status int, reason *string, operatorID *int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE teacher_applications SET status = $2, reason = $3, operator_id = $4, updated_at = NOW() WHERE id = $1`,
		id, status, reason, operatorID,
	)
	return err
}

// TeacherRanking

func (r *Repository) GetTeacherRanking(ctx context.Context, limit int) ([]*TeacherRanking, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, name, avatar, rating, order_count, balance
		 FROM teachers WHERE status = 1 ORDER BY order_count DESC, rating DESC LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*TeacherRanking
	rank := 1
	for rows.Next() {
		var item TeacherRanking
		if err := rows.Scan(&item.TeacherID, &item.Name, &item.Avatar, &item.Rating, &item.OrderCount, &item.Balance); err != nil {
			return nil, err
		}
		item.Rank = rank
		rank++
		items = append(items, &item)
	}
	return items, nil
}

// TeacherDashboardStats

func (r *Repository) GetTeacherDashboardStats(ctx context.Context, teacherID int64) (*TeacherDashboardStats, error) {
	// Get teacher basic info
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT rating, order_count, balance FROM teachers WHERE id = $1`,
		teacherID,
	)
	var stats TeacherDashboardStats
	var balance float64
	if err := row.Scan(&stats.Rating, &stats.TotalOrders, &balance); err != nil {
		return nil, fmt.Errorf("get teacher stats: %w", err)
	}
	stats.TotalIncome = balance

	// Get today's balance logs sum
	row = r.dbtx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM teacher_balance_logs WHERE teacher_id = $1 AND change_type = 'income' AND created_at >= $2`,
		teacherID, time.Now().Truncate(24*time.Hour),
	)
	var todayIncome float64
	if err := row.Scan(&todayIncome); err != nil {
		todayIncome = 0
	}
	stats.TodayIncome = todayIncome

	// Get this week's balance logs sum
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	row = r.dbtx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM teacher_balance_logs WHERE teacher_id = $1 AND change_type = 'income' AND created_at >= $2`,
		teacherID, weekStart,
	)
	var weekIncome float64
	if err := row.Scan(&weekIncome); err != nil {
		weekIncome = 0
	}
	stats.WeekIncome = weekIncome

	// Get this month's balance logs sum
	monthStart := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	row = r.dbtx.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM teacher_balance_logs WHERE teacher_id = $1 AND change_type = 'income' AND created_at >= $2`,
		teacherID, monthStart,
	)
	var monthIncome float64
	if err := row.Scan(&monthIncome); err != nil {
		monthIncome = 0
	}
	stats.MonthIncome = monthIncome

	// Completion rate (simplified)
	if stats.TotalOrders > 0 {
		stats.CompletionRate = 100.0
	}

	return &stats, nil
}

// TeacherLevelGoods

func (r *Repository) GetTeacherLevelGoods(ctx context.Context, levelID int64) ([]*TeacherLevelGoods, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, level_id, goods_id, commission_rate, created_at FROM teacher_level_goods WHERE level_id = $1`,
		levelID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*TeacherLevelGoods
	for rows.Next() {
		var item TeacherLevelGoods
		if err := rows.Scan(&item.ID, &item.LevelID, &item.GoodsID, &item.CommissionRate, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

func (r *Repository) UpdateTeacherLevelGoods(ctx context.Context, levelID int64, goods []*TeacherLevelGoods) error {
	// Delete existing goods for this level
	if _, err := r.dbtx.ExecContext(ctx, `DELETE FROM teacher_level_goods WHERE level_id = $1`, levelID); err != nil {
		return err
	}

	// Insert new goods
	for _, g := range goods {
		if _, err := r.dbtx.ExecContext(ctx,
			`INSERT INTO teacher_level_goods (level_id, goods_id, commission_rate) VALUES ($1, $2, $3)`,
			levelID, g.GoodsID, g.CommissionRate,
		); err != nil {
			return err
		}
	}

	return nil
}
