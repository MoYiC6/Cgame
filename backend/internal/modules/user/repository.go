package user

import (
	"context"
	"database/sql"
	"encoding/json"

	"backend/internal/platform/database"
)

type Repository interface {
	GetByID(ctx context.Context, userID int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	CreateBalanceLog(ctx context.Context, log *UserBalanceLog) error
	GetUserBalanceLogs(ctx context.Context, userID int64, page, pageSize int) ([]*UserBalanceLog, int, error)
	GetUserLevel(ctx context.Context, userID int64) (*UserLevel, error)
	CreateUserLevelLog(ctx context.Context, log *UserLevelLog) error
	CreatePurchaseRecord(ctx context.Context, record *UserPurchaseRecord) error
	GetUserPurchaseCount(ctx context.Context, userID, goodsID int64) (int, error)
}

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

func (r *repository) GetByID(ctx context.Context, userID int64) (*User, error) {
	row := r.executor(ctx).QueryRowContext(ctx, `
		SELECT id, public_id, email, password_hash, status, password_changed_at, last_login_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`, userID)
	return scanUser(row)
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.executor(ctx).QueryRowContext(ctx, `
		SELECT id, public_id, email, password_hash, status, password_changed_at, last_login_at, created_at, updated_at
		FROM users
		WHERE email = $1
	`, NormalizeEmail(email))
	return scanUser(row)
}

func (r *repository) executor(ctx context.Context) database.DBTX {
	return database.ExecutorFromContext(ctx, r.dbtx)
}

func scanUser(row rowScanner) (*User, error) {
	if row == nil {
		return nil, nil
	}
	var user User
	err := row.Scan(
		&user.ID,
		&user.PublicID,
		&user.Email,
		&user.PasswordHash,
		&user.Status,
		&user.PasswordChangedAt,
		&user.LastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func (r *repository) CreateBalanceLog(ctx context.Context, log *UserBalanceLog) error {
	return r.executor(ctx).QueryRowContext(ctx,
		`INSERT INTO user_balance_logs (user_id, change_type, amount, balance_after, related_id, related_no, description)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		log.UserID, log.ChangeType, log.Amount, log.BalanceAfter, log.RelatedID, log.RelatedNo, log.Description,
	).Scan(&log.ID)
}

func (r *repository) GetUserBalanceLogs(ctx context.Context, userID int64, page, pageSize int) ([]*UserBalanceLog, int, error) {
	countQuery := "SELECT COUNT(*) FROM user_balance_logs WHERE user_id = $1"
	var total int
	if err := r.executor(ctx).QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.executor(ctx).QueryContext(ctx,
		`SELECT id, user_id, change_type, amount, balance_after, related_id, related_no, description, created_at
		 FROM user_balance_logs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*UserBalanceLog
	for rows.Next() {
		var l UserBalanceLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.ChangeType, &l.Amount, &l.BalanceAfter, &l.RelatedID, &l.RelatedNo, &l.Description, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		logs = append(logs, &l)
	}
	return logs, total, nil
}

func (r *repository) GetUserLevel(ctx context.Context, userID int64) (*UserLevel, error) {
	row := r.executor(ctx).QueryRowContext(ctx,
		`SELECT id, name, min_consumption, discount_rate, benefits, status, created_at, updated_at FROM user_levels WHERE status = 1 LIMIT 1`,
	)
	var l UserLevel
	var benefitsJSON []byte
	err := row.Scan(&l.ID, &l.Name, &l.MinConsumption, &l.DiscountRate, &benefitsJSON, &l.Status, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, nil
	}
	if len(benefitsJSON) > 0 {
		json.Unmarshal(benefitsJSON, &l.Benefits)
	}
	return &l, nil
}

func (r *repository) CreateUserLevelLog(ctx context.Context, log *UserLevelLog) error {
	return r.executor(ctx).QueryRowContext(ctx,
		`INSERT INTO user_level_logs (user_id, old_level_id, new_level_id, change_reason) VALUES ($1, $2, $3, $4) RETURNING id`,
		log.UserID, log.OldLevelID, log.NewLevelID, log.ChangeReason,
	).Scan(&log.ID)
}

func (r *repository) CreatePurchaseRecord(ctx context.Context, record *UserPurchaseRecord) error {
	return r.executor(ctx).QueryRowContext(ctx,
		`INSERT INTO user_purchase_records (user_id, goods_id, order_id, quantity) VALUES ($1, $2, $3, $4) RETURNING id`,
		record.UserID, record.GoodsID, record.OrderID, record.Quantity,
	).Scan(&record.ID)
}

func (r *repository) GetUserPurchaseCount(ctx context.Context, userID, goodsID int64) (int, error) {
	var count int
	err := r.executor(ctx).QueryRowContext(ctx,
		"SELECT COALESCE(SUM(quantity), 0) FROM user_purchase_records WHERE user_id = $1 AND goods_id = $2",
		userID, goodsID,
	).Scan(&count)
	return count, err
}
