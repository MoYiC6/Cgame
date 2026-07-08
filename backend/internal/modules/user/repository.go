package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"backend/internal/platform/database"
)

type Repository interface {
	GetByID(ctx context.Context, userID int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByIdentifier(ctx context.Context, identifier string) (*User, error)
	ListUsers(ctx context.Context, query UserQuery) ([]*User, int, error)
	UpdateUser(ctx context.Context, user *User) error
	UpdateUserStatus(ctx context.Context, userID int64, status int16) error
	GetUserLevelByID(ctx context.Context, levelID int64) (*UserLevel, error)
	GetUserLevels(ctx context.Context) ([]*UserLevel, error)
	CreateBalanceLog(ctx context.Context, log *UserBalanceLog) error
	GetUserBalanceLogs(ctx context.Context, userID int64, page, pageSize int) ([]*UserBalanceLog, int, error)
	GetRecentBalanceLogs(ctx context.Context, userID int64, limit int) ([]*UserBalanceLog, error)
	GetUserLevel(ctx context.Context, userID int64) (*UserLevel, error)
	CreateUserLevelLog(ctx context.Context, log *UserLevelLog) error
	CreatePurchaseRecord(ctx context.Context, record *UserPurchaseRecord) error
	GetUserPurchaseCount(ctx context.Context, userID, goodsID int64) (int, error)
	GetConsumptionRanking(ctx context.Context, limit int) ([]*ConsumptionRankingItem, error)
	ListUserSelectors(ctx context.Context, keyword string, limit int) ([]*UserSelectorItem, error)
	ListUserLoginLogs(ctx context.Context, userID *int64, page, pageSize int) ([]*UserLoginLog, int, error)
	DeleteUserLoginLogs(ctx context.Context, ids []int64) error
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

func (r *repository) ListUsers(ctx context.Context, query UserQuery) ([]*User, int, error) {
	where := "WHERE deleted = 0"
	var args []any
	argIdx := 1

	if query.Username != "" {
		where += fmt.Sprintf(" AND username ILIKE $%d", argIdx)
		args = append(args, "%"+query.Username+"%")
		argIdx++
	}
	if query.Nickname != "" {
		where += fmt.Sprintf(" AND nickname ILIKE $%d", argIdx)
		args = append(args, "%"+query.Nickname+"%")
		argIdx++
	}
	if query.Mobile != "" {
		where += fmt.Sprintf(" AND mobile ILIKE $%d", argIdx)
		args = append(args, "%"+query.Mobile+"%")
		argIdx++
	}
	if query.Email != "" {
		where += fmt.Sprintf(" AND email ILIKE $%d", argIdx)
		args = append(args, "%"+query.Email+"%")
		argIdx++
	}
	if query.Status != nil {
		where += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, *query.Status)
		argIdx++
	}
	if query.IsTeacher != nil {
		where += fmt.Sprintf(" AND is_teacher = $%d", argIdx)
		args = append(args, *query.IsTeacher)
		argIdx++
	}
	if query.LevelID != nil {
		where += fmt.Sprintf(" AND level_id = $%d", argIdx)
		args = append(args, *query.LevelID)
		argIdx++
	}
	if query.StartTime != nil {
		where += fmt.Sprintf(" AND create_time >= $%d", argIdx)
		args = append(args, *query.StartTime)
		argIdx++
	}
	if query.EndTime != nil {
		where += fmt.Sprintf(" AND create_time <= $%d", argIdx)
		args = append(args, *query.EndTime)
		argIdx++
	}

	countQuery := "SELECT COUNT(*) FROM sys_user " + where
	var total int
	if err := r.executor(ctx).QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf(`SELECT id, username, email, password, nickname, real_name, mobile, email_verified, avatar, gender, birthday, intro, province, city, district, wechat, status, is_teacher, level_id, last_login_time, create_time, update_time
		 FROM sys_user %s ORDER BY id DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, query.PageSize, (query.PageNum-1)*query.PageSize)

	rows, err := r.executor(ctx).QueryContext(ctx, querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *repository) UpdateUser(ctx context.Context, user *User) error {
	_, err := r.executor(ctx).ExecContext(ctx, `
		UPDATE sys_user SET
			nickname = $1, avatar = $2, gender = $3, birthday = $4,
			intro = $5, province = $6, city = $7, district = $8,
			mobile = $9, email = $10, real_name = $11,
			update_time = NOW()
		WHERE id = $12 AND deleted = 0
	`, user.Nickname, user.Avatar, user.Gender, user.Birthday,
		user.Intro, user.Province, user.City, user.District,
		user.Mobile, user.Email, user.RealName, user.ID)
	return err
}

func (r *repository) UpdateUserStatus(ctx context.Context, userID int64, status int16) error {
	_, err := r.executor(ctx).ExecContext(ctx,
		"UPDATE sys_user SET status = $1, update_time = NOW() WHERE id = $2 AND deleted = 0",
		status, userID)
	return err
}

func (r *repository) GetUserLevelByID(ctx context.Context, levelID int64) (*UserLevel, error) {
	row := r.executor(ctx).QueryRowContext(ctx,
		`SELECT id, name, min_consumption, discount_rate, benefits, status, created_at, updated_at FROM user_levels WHERE id = $1`,
		levelID)
	var l UserLevel
	var benefitsJSON []byte
	var status sql.NullInt16
	err := row.Scan(&l.ID, &l.Name, &l.MinConsumption, &l.DiscountRate, &benefitsJSON, &status, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if status.Valid {
		v := int(status.Int16)
		l.Status = &v
	}
	if len(benefitsJSON) > 0 {
		json.Unmarshal(benefitsJSON, &l.Benefits)
	}
	return &l, nil
}

func (r *repository) GetUserLevels(ctx context.Context) ([]*UserLevel, error) {
	rows, err := r.executor(ctx).QueryContext(ctx,
		`SELECT id, name, min_consumption, discount_rate, benefits, status, created_at, updated_at FROM user_levels WHERE status = 1 ORDER BY min_consumption ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var levels []*UserLevel
	for rows.Next() {
		var l UserLevel
		var benefitsJSON []byte
		var status sql.NullInt16
		if err := rows.Scan(&l.ID, &l.Name, &l.MinConsumption, &l.DiscountRate, &benefitsJSON, &status, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		if status.Valid {
			v := int(status.Int16)
			l.Status = &v
		}
		if len(benefitsJSON) > 0 {
			json.Unmarshal(benefitsJSON, &l.Benefits)
		}
		levels = append(levels, &l)
	}
	return levels, nil
}

func (r *repository) GetRecentBalanceLogs(ctx context.Context, userID int64, limit int) ([]*UserBalanceLog, error) {
	rows, err := r.executor(ctx).QueryContext(ctx,
		`SELECT id, user_id, change_type, amount, balance_after, related_id, related_no, description, created_at
		 FROM user_balance_logs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*UserBalanceLog
	for rows.Next() {
		var l UserBalanceLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.ChangeType, &l.Amount, &l.BalanceAfter, &l.RelatedID, &l.RelatedNo, &l.Description, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, &l)
	}
	return logs, nil
}

func (r *repository) GetConsumptionRanking(ctx context.Context, limit int) ([]*ConsumptionRankingItem, error) {
	rows, err := r.executor(ctx).QueryContext(ctx,
		`SELECT id, username, nickname, avatar, total_consumption
		 FROM sys_user WHERE deleted = 0 AND status = 1 ORDER BY total_consumption DESC LIMIT $1`,
		limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*ConsumptionRankingItem
	rank := 1
	for rows.Next() {
		var item ConsumptionRankingItem
		if err := rows.Scan(&item.UserID, &item.Username, &item.Nickname, &item.Avatar, &item.Consumption); err != nil {
			return nil, err
		}
		item.Rank = rank
		rank++
		items = append(items, &item)
	}
	return items, nil
}

func (r *repository) ListUserSelectors(ctx context.Context, keyword string, limit int) ([]*UserSelectorItem, error) {
	where := "WHERE deleted = 0"
	var args []any
	if keyword != "" {
		where += " AND (username ILIKE $1 OR nickname ILIKE $1 OR mobile ILIKE $1)"
		args = append(args, "%"+keyword+"%")
	}
	querySQL := fmt.Sprintf("SELECT id, username, nickname, mobile, avatar FROM sys_user %s ORDER BY id DESC LIMIT $%d", where, len(args)+1)
	args = append(args, limit)

	rows, err := r.executor(ctx).QueryContext(ctx, querySQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*UserSelectorItem
	for rows.Next() {
		var item UserSelectorItem
		if err := rows.Scan(&item.ID, &item.Username, &item.Nickname, &item.Mobile, &item.Avatar); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	return items, nil
}

func (r *repository) ListUserLoginLogs(ctx context.Context, userID *int64, page, pageSize int) ([]*UserLoginLog, int, error) {
	where := "WHERE 1=1"
	var args []any
	argIdx := 1
	if userID != nil {
		where += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *userID)
		argIdx++
	}

	countQuery := "SELECT COUNT(*) FROM user_login_logs " + where
	var total int
	if err := r.executor(ctx).QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	querySQL := fmt.Sprintf(`SELECT id, user_id, login_type, ip_address, user_agent, login_status, fail_reason, created_at
		 FROM user_login_logs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.executor(ctx).QueryContext(ctx, querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*UserLoginLog
	for rows.Next() {
		var l UserLoginLog
		var userIDPtr sql.NullInt64
		if err := rows.Scan(&l.ID, &userIDPtr, &l.LoginType, &l.IPAddress, &l.UserAgent, &l.LoginStatus, &l.FailReason, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		if userIDPtr.Valid {
			l.UserID = &userIDPtr.Int64
		}
		logs = append(logs, &l)
	}
	return logs, total, nil
}

func (r *repository) DeleteUserLoginLogs(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	_, err := r.executor(ctx).ExecContext(ctx,
		fmt.Sprintf("DELETE FROM user_login_logs WHERE id IN (%s)", strings.Join(placeholders, ",")),
		args...)
	return err
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
