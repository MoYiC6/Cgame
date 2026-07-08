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
	GetByIdentifier(ctx context.Context, identifier string) (*User, error)
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
		SELECT id, username, email, password, nickname, real_name, mobile, email_verified, avatar, gender, birthday, intro, province, city, district, wechat, status, is_teacher, level_id, last_login_time, create_time, update_time
		FROM sys_user
		WHERE id = $1 AND deleted = 0
	`, userID)
	return scanUser(row)
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.executor(ctx).QueryRowContext(ctx, `
		SELECT id, username, email, password, nickname, real_name, mobile, email_verified, avatar, gender, birthday, intro, province, city, district, wechat, status, is_teacher, level_id, last_login_time, create_time, update_time
		FROM sys_user
		WHERE email = $1 AND deleted = 0
	`, NormalizeEmail(email))
	return scanUser(row)
}

func (r *repository) GetByIdentifier(ctx context.Context, identifier string) (*User, error) {
	row := r.executor(ctx).QueryRowContext(ctx, `
		SELECT id, username, email, password, nickname, real_name, mobile, email_verified, avatar, gender, birthday, intro, province, city, district, wechat, status, is_teacher, level_id, last_login_time, create_time, update_time
		FROM sys_user
		WHERE (username = $1 OR mobile = $1) AND deleted = 0
		LIMIT 1
	`, identifier)
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
	var birthday sql.NullTime
	var lastLoginTime sql.NullTime
	var levelID sql.NullInt64
	var statusInt sql.NullInt16
	var realName sql.NullString
	var mobile sql.NullString
	var email sql.NullString
	var passwordHash sql.NullString
	var avatar sql.NullString
	var intro sql.NullString
	var province sql.NullString
	var city sql.NullString
	var district sql.NullString
	var wechat sql.NullString
	err := row.Scan(
		&user.ID,
		&user.Username,
		&email,
		&passwordHash,
		&user.Nickname,
		&realName,
		&mobile,
		&user.EmailVerified,
		&avatar,
		&user.Gender,
		&birthday,
		&intro,
		&province,
		&city,
		&district,
		&wechat,
		&statusInt,
		&user.IsTeacher,
		&levelID,
		&lastLoginTime,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	user.Email = email.String
	user.PasswordHash = passwordHash.String
	user.RealName = realName.String
	user.Mobile = mobile.String
	user.Avatar = avatar.String
	user.Intro = intro.String
	user.Province = province.String
	user.City = city.String
	user.District = district.String
	user.Wechat = wechat.String
	if birthday.Valid {
		user.Birthday = &birthday.Time
	}
	if lastLoginTime.Valid {
		user.LastLoginTime = &lastLoginTime.Time
	}
	if levelID.Valid {
		user.LevelID = &levelID.Int64
	}
	user.Status = mapStatus(statusInt)
	user.PublicID = user.Username
	return &user, nil
}

func mapStatus(v sql.NullInt16) string {
	if !v.Valid {
		return "disabled"
	}
	switch v.Int16 {
	case 1:
		return "active"
	case 2:
		return "locked"
	default:
		return "disabled"
	}
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
