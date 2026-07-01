package teacher

import (
	"context"
	"encoding/json"
	"fmt"

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

func (r *Repository) CreateTeacherStatusLog(ctx context.Context, log *TeacherStatusLog) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO teacher_status_logs (teacher_id, old_status, new_status, reason, operator_id) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		log.TeacherID, log.OldStatus, log.NewStatus, log.Reason, log.OperatorID,
	).Scan(&log.ID)
}

func (r *Repository) CreateTeacherBalanceLog(ctx context.Context, log *TeacherBalanceLog) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO teacher_balance_logs (teacher_id, change_type, amount, balance_after, related_id, related_no, description) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		log.TeacherID, log.ChangeType, log.Amount, log.BalanceAfter, log.RelatedID, log.RelatedNo, log.Description,
	).Scan(&log.ID)
}
