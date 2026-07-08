package invite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/internal/platform/database"
)

type Repository interface {
	// Client invite operations
	GetInviteInfo(ctx context.Context, userID int64) (*InviteInfo, error)
	ListInviteRecords(ctx context.Context, userID int64, page, pageSize int) (*InviteRecordPageResult, error)
	CreateInviteRecord(ctx context.Context, inviterID, inviteeID int64, inviteCode string) error
	GetInviteRecordByInvitee(ctx context.Context, inviteeID int64) (*InviteRecord, error)

	// Teacher invite code operations
	GetTeacherInviteCodeByUser(ctx context.Context, userID int64) (*TeacherInviteCode, error)
	CreateTeacherInviteCode(ctx context.Context, userID int64, code string) error
	GetTeacherInviteCodeByCode(ctx context.Context, code string) (*TeacherInviteCode, error)

	// Admin invite code operations
	ListTeacherInviteCodes(ctx context.Context, query InviteCodeQuery) (*TeacherInviteCodePageResult, error)
	CreateTeacherInviteCodes(ctx context.Context, codes []TeacherInviteCode) error
	GetTeacherInviteCodeByID(ctx context.Context, id int64) (*TeacherInviteCode, error)
	UpdateTeacherInviteCode(ctx context.Context, id int64, updates map[string]any) error
	DeleteTeacherInviteCode(ctx context.Context, id int64) error
	CountTeacherInviteCodes(ctx context.Context, query InviteCodeQuery) (int64, error)
}

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

// Client invite operations

func (r *repository) GetInviteInfo(ctx context.Context, userID int64) (*InviteInfo, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var totalInvited, rewardedCount int
	var inviteCode string

	// Get total invited count
	if err := exec.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM invite_records WHERE inviter_id = $1`,
		userID,
	).Scan(&totalInvited); err != nil {
		return nil, fmt.Errorf("count invite records: %w", err)
	}

	// Get rewarded count (those with reward coupon)
	if err := exec.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM invite_records WHERE inviter_id = $1 AND inviter_reward_coupon_id IS NOT NULL`,
		userID,
	).Scan(&rewardedCount); err != nil {
		return nil, fmt.Errorf("count rewarded invites: %w", err)
	}

	// Get user's own invite code (from teacher invite codes if they have one)
	row := exec.QueryRowContext(ctx,
		`SELECT code FROM teacher_invite_codes WHERE used_by = $1 AND status = 'used' LIMIT 1`,
		userID,
	)
	_ = row.Scan(&inviteCode) // may be empty

	return &InviteInfo{
		InviteCode:              inviteCode,
		InviteLink:              buildInviteLink("", inviteCode),
		TotalInvited:            totalInvited,
		RewardedCount:           rewardedCount,
		InviterRewardCouponName: "",
		InviteeRewardCouponName: "",
	}, nil
}

func (r *repository) ListInviteRecords(ctx context.Context, userID int64, page, pageSize int) (*InviteRecordPageResult, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var total int64
	if err := exec.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM invite_records WHERE inviter_id = $1`,
		userID,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("count invite records: %w", err)
	}

	rows, err := exec.QueryContext(ctx,
		`SELECT id, inviter_id, invitee_id, invite_code, inviter_reward_coupon_id, invitee_reward_coupon_id, created_at
		 FROM invite_records
		 WHERE inviter_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, fmt.Errorf("list invite records: %w", err)
	}
	defer rows.Close()

	records := []InviteRecord{}
	for rows.Next() {
		var record InviteRecord
		if err := rows.Scan(&record.ID, &record.InviterID, &record.InviteeID, &record.InviteCode,
			&record.InviterRewardCouponID, &record.InviteeRewardCouponID, &record.CreateTime); err != nil {
			return nil, fmt.Errorf("scan invite record: %w", err)
		}
		records = append(records, record)
	}

	return &InviteRecordPageResult{
		Total:    total,
		PageNum:  page,
		PageSize: pageSize,
		Records:  records,
	}, nil
}

func (r *repository) CreateInviteRecord(ctx context.Context, inviterID, inviteeID int64, inviteCode string) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`INSERT INTO invite_records (inviter_id, invitee_id, invite_code, created_at)
		 VALUES ($1, $2, $3, NOW())`,
		inviterID, inviteeID, inviteCode,
	)
	if err != nil {
		return fmt.Errorf("create invite record: %w", err)
	}
	return nil
}

func (r *repository) GetInviteRecordByInvitee(ctx context.Context, inviteeID int64) (*InviteRecord, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var record InviteRecord
	if err := exec.QueryRowContext(ctx,
		`SELECT id, inviter_id, invitee_id, invite_code, inviter_reward_coupon_id, invitee_reward_coupon_id, created_at
		 FROM invite_records WHERE invitee_id = $1`,
		inviteeID,
	).Scan(&record.ID, &record.InviterID, &record.InviteeID, &record.InviteCode,
		&record.InviterRewardCouponID, &record.InviteeRewardCouponID, &record.CreateTime); err != nil {
		return nil, fmt.Errorf("get invite record by invitee: %w", err)
	}
	return &record, nil
}

// Teacher invite code operations

func (r *repository) GetTeacherInviteCodeByUser(ctx context.Context, userID int64) (*TeacherInviteCode, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var code TeacherInviteCode
	if err := exec.QueryRowContext(ctx,
		`SELECT id, code, status, remark, expire_time, created_by, created_at, used_by, used_time, teacher_id, revoked_by, revoked_time, updated_at
		 FROM teacher_invite_codes WHERE used_by = $1`,
		userID,
	).Scan(&code.ID, &code.Code, &code.Status, &code.Remark, &code.ExpireTime, &code.CreatedBy, &code.CreateTime,
		&code.UsedBy, &code.UsedTime, &code.TeacherID, &code.RevokedBy, &code.RevokedTime, &code.UpdateTime); err != nil {
		return nil, fmt.Errorf("get teacher invite code by user: %w", err)
	}
	return &code, nil
}

func (r *repository) CreateTeacherInviteCode(ctx context.Context, userID int64, code string) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`INSERT INTO teacher_invite_codes (code, status, created_by, created_at, updated_at)
		 VALUES ($1, 'unused', $2, NOW(), NOW())`,
		code, userID,
	)
	if err != nil {
		return fmt.Errorf("create teacher invite code: %w", err)
	}
	return nil
}

func (r *repository) GetTeacherInviteCodeByCode(ctx context.Context, code string) (*TeacherInviteCode, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var result TeacherInviteCode
	if err := exec.QueryRowContext(ctx,
		`SELECT id, code, status, remark, expire_time, created_by, created_at, used_by, used_time, teacher_id, revoked_by, revoked_time, updated_at
		 FROM teacher_invite_codes WHERE code = $1`,
		code,
	).Scan(&result.ID, &result.Code, &result.Status, &result.Remark, &result.ExpireTime, &result.CreatedBy, &result.CreateTime,
		&result.UsedBy, &result.UsedTime, &result.TeacherID, &result.RevokedBy, &result.RevokedTime, &result.UpdateTime); err != nil {
		return nil, fmt.Errorf("get teacher invite code by code: %w", err)
	}
	return &result, nil
}

// Admin invite code operations

func (r *repository) ListTeacherInviteCodes(ctx context.Context, query InviteCodeQuery) (*TeacherInviteCodePageResult, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where, args := buildAdminWhere(query)

	var total int64
	if err := exec.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM teacher_invite_codes "+where, args...,
	).Scan(&total); err != nil {
		return nil, fmt.Errorf("count teacher invite codes: %w", err)
	}

	queryArgs := append([]any{}, args...)
	limitIndex := len(queryArgs) + 1
	offsetIndex := len(queryArgs) + 2
	queryArgs = append(queryArgs, query.PageSize, (query.PageNum-1)*query.PageSize)

	rows, err := exec.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, code, status, remark, expire_time, created_by, created_at, used_by, used_time, teacher_id, revoked_by, revoked_time, updated_at
		 FROM teacher_invite_codes %s
		 ORDER BY created_at DESC
		 LIMIT $%d OFFSET $%d`, where, limitIndex, offsetIndex),
		queryArgs...,
	)
	if err != nil {
		return nil, fmt.Errorf("list teacher invite codes: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	records := []TeacherInviteCodeVO{}
	for rows.Next() {
		var code TeacherInviteCode
		if err := rows.Scan(&code.ID, &code.Code, &code.Status, &code.Remark, &code.ExpireTime, &code.CreatedBy, &code.CreateTime,
			&code.UsedBy, &code.UsedTime, &code.TeacherID, &code.RevokedBy, &code.RevokedTime, &code.UpdateTime); err != nil {
			return nil, fmt.Errorf("scan teacher invite code: %w", err)
		}
		records = append(records, toTeacherInviteCodeVO(code, now))
	}

	return &TeacherInviteCodePageResult{
		Total:    total,
		PageNum:  query.PageNum,
		PageSize: query.PageSize,
		Records:  records,
	}, nil
}

func (r *repository) CreateTeacherInviteCodes(ctx context.Context, codes []TeacherInviteCode) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	for _, code := range codes {
		_, err := exec.ExecContext(ctx,
			`INSERT INTO teacher_invite_codes (code, status, remark, expire_time, created_by, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`,
			code.Code, code.Status, code.Remark, code.ExpireTime, code.CreatedBy,
		)
		if err != nil {
			return fmt.Errorf("create teacher invite code %s: %w", code.Code, err)
		}
	}
	return nil
}

func (r *repository) GetTeacherInviteCodeByID(ctx context.Context, id int64) (*TeacherInviteCode, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var code TeacherInviteCode
	if err := exec.QueryRowContext(ctx,
		`SELECT id, code, status, remark, expire_time, created_by, created_at, used_by, used_time, teacher_id, revoked_by, revoked_time, updated_at
		 FROM teacher_invite_codes WHERE id = $1`,
		id,
	).Scan(&code.ID, &code.Code, &code.Status, &code.Remark, &code.ExpireTime, &code.CreatedBy, &code.CreateTime,
		&code.UsedBy, &code.UsedTime, &code.TeacherID, &code.RevokedBy, &code.RevokedTime, &code.UpdateTime); err != nil {
		return nil, fmt.Errorf("get teacher invite code by id: %w", err)
	}
	return &code, nil
}

func (r *repository) UpdateTeacherInviteCode(ctx context.Context, id int64, updates map[string]any) error {
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
	query := fmt.Sprintf("UPDATE teacher_invite_codes SET %s, updated_at = NOW() WHERE id = $%d",
		strings.Join(sets, ", "), len(args))
	_, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update teacher invite code: %w", err)
	}
	return nil
}

func (r *repository) DeleteTeacherInviteCode(ctx context.Context, id int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx, `DELETE FROM teacher_invite_codes WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete teacher invite code: %w", err)
	}
	return nil
}

func (r *repository) CountTeacherInviteCodes(ctx context.Context, query InviteCodeQuery) (int64, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where, args := buildAdminWhere(query)
	var total int64
	if err := exec.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM teacher_invite_codes "+where, args...,
	).Scan(&total); err != nil {
		return 0, fmt.Errorf("count teacher invite codes: %w", err)
	}
	return total, nil
}

func buildAdminWhere(query InviteCodeQuery) (string, []any) {
	where := "WHERE 1=1"
	args := []any{}

	if query.Code != "" {
		args = append(args, "%"+query.Code+"%")
		where += fmt.Sprintf(" AND code ILIKE $%d", len(args))
	}
	if query.Status != "" {
		args = append(args, query.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if query.UsedBy != nil {
		args = append(args, *query.UsedBy)
		where += fmt.Sprintf(" AND used_by = $%d", len(args))
	}
	if query.CreatedBy != nil {
		args = append(args, *query.CreatedBy)
		where += fmt.Sprintf(" AND created_by = $%d", len(args))
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

func toTeacherInviteCodeVO(code TeacherInviteCode, now time.Time) TeacherInviteCodeVO {
	status := resolveStatus(code, now)
	return TeacherInviteCodeVO{
		ID:          code.ID,
		Code:        code.Code,
		Status:      status,
		StatusDesc:  statusDesc(status),
		Remark:      code.Remark,
		ExpireTime:  code.ExpireTime,
		CreatedBy:   code.CreatedBy,
		CreateTime:  code.CreateTime,
		UsedBy:      code.UsedBy,
		UsedTime:    code.UsedTime,
		TeacherID:   code.TeacherID,
		RevokedBy:   code.RevokedBy,
		RevokedTime: code.RevokedTime,
	}
}

// InviteRecordPageResult wraps paginated invite records.
type InviteRecordPageResult struct {
	Total    int64           `json:"total"`
	PageNum  int             `json:"pageNum"`
	PageSize int             `json:"pageSize"`
	Records  []InviteRecord  `json:"records"`
}
