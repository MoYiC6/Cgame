package partner

import (
	"context"
	"fmt"
	"strings"

	"backend/internal/platform/database"
)

type Repository interface {
	// Partner config operations
	CreatePartnerConfig(ctx context.Context, config *PartnerConfig) error
	GetPartnerConfigByID(ctx context.Context, id int64) (*PartnerConfig, error)
	ListPartnerConfigs(ctx context.Context, query PartnerConfigQuery) (*PartnerConfigPageResult, error)
	UpdatePartnerConfig(ctx context.Context, id int64, updates map[string]any) error
	DeletePartnerConfig(ctx context.Context, id int64) error

	// Teacher partner operations
	CreateTeacherPartner(ctx context.Context, tp *TeacherPartner) error
	GetTeacherPartnerByID(ctx context.Context, id int64) (*TeacherPartner, error)
	ListTeacherPartners(ctx context.Context, query TeacherPartnerQuery) (*TeacherPartnerPageResult, error)
	UpdateTeacherPartner(ctx context.Context, id int64, updates map[string]any) error
	DeleteTeacherPartner(ctx context.Context, id int64) error
	ListPartneredTeachers(ctx context.Context, page, pageSize int) (*TeacherPartnerPageResult, error)
}

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

// Partner config operations

func (r *repository) CreatePartnerConfig(ctx context.Context, config *PartnerConfig) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO partner_configs (name, partner_type, commission_rate, fixed_fee, description, contact_name, contact_phone, contact_email, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) RETURNING id`,
		config.Name, config.PartnerType, config.CommissionRate, config.FixedFee, config.Description,
		config.ContactName, config.ContactPhone, config.ContactEmail, config.Status,
	).Scan(&config.ID)
}

func (r *repository) GetPartnerConfigByID(ctx context.Context, id int64) (*PartnerConfig, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var p PartnerConfig
	if err := exec.QueryRowContext(ctx,
		`SELECT id, name, partner_type, commission_rate, fixed_fee, description, contact_name, contact_phone, contact_email, status, created_at, updated_at
		 FROM partner_configs WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &p.PartnerType, &p.CommissionRate, &p.FixedFee, &p.Description,
		&p.ContactName, &p.ContactPhone, &p.ContactEmail, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get partner config by id: %w", err)
	}
	return &p, nil
}

func (r *repository) ListPartnerConfigs(ctx context.Context, query PartnerConfigQuery) (*PartnerConfigPageResult, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where, args := buildPartnerConfigWhere(query)

	var total int64
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM partner_configs "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count partner configs: %w", err)
	}

	queryArgs := append([]any{}, args...)
	limitIndex := len(queryArgs) + 1
	offsetIndex := len(queryArgs) + 2
	queryArgs = append(queryArgs, query.PageSize, (query.PageNum-1)*query.PageSize)

	rows, err := exec.QueryContext(ctx,
		fmt.Sprintf(`SELECT id, name, partner_type, commission_rate, fixed_fee, description, contact_name, contact_phone, contact_email, status, created_at, updated_at
		 FROM partner_configs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, limitIndex, offsetIndex),
		queryArgs...,
	)
	if err != nil {
		return nil, fmt.Errorf("list partner configs: %w", err)
	}
	defer rows.Close()

	records := []PartnerConfigVO{}
	for rows.Next() {
		var p PartnerConfig
		if err := rows.Scan(&p.ID, &p.Name, &p.PartnerType, &p.CommissionRate, &p.FixedFee, &p.Description,
			&p.ContactName, &p.ContactPhone, &p.ContactEmail, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan partner config: %w", err)
		}
		records = append(records, toPartnerConfigVO(p))
	}
	return &PartnerConfigPageResult{Total: total, PageNum: query.PageNum, PageSize: query.PageSize, Records: records}, nil
}

func (r *repository) UpdatePartnerConfig(ctx context.Context, id int64, updates map[string]any) error {
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
	query := fmt.Sprintf("UPDATE partner_configs SET %s, updated_at = NOW() WHERE id = $%d", strings.Join(sets, ", "), len(args))
	_, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update partner config: %w", err)
	}
	return nil
}

func (r *repository) DeletePartnerConfig(ctx context.Context, id int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx, `DELETE FROM partner_configs WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete partner config: %w", err)
	}
	return nil
}

// Teacher partner operations

func (r *repository) CreateTeacherPartner(ctx context.Context, tp *TeacherPartner) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO teacher_partners (teacher_id, partner_id, partner_config_id, cooperation_type, commission_rate, start_date, end_date, status, remark, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) RETURNING id`,
		tp.TeacherID, tp.PartnerID, tp.PartnerConfigID, tp.CooperationType, tp.CommissionRate,
		tp.StartDate, tp.EndDate, tp.Status, tp.Remark,
	).Scan(&tp.ID)
}

func (r *repository) GetTeacherPartnerByID(ctx context.Context, id int64) (*TeacherPartner, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var tp TeacherPartner
	if err := exec.QueryRowContext(ctx,
		`SELECT id, teacher_id, partner_id, partner_config_id, cooperation_type, commission_rate, start_date, end_date, status, remark, created_at, updated_at
		 FROM teacher_partners WHERE id = $1`, id,
	).Scan(&tp.ID, &tp.TeacherID, &tp.PartnerID, &tp.PartnerConfigID, &tp.CooperationType, &tp.CommissionRate,
		&tp.StartDate, &tp.EndDate, &tp.Status, &tp.Remark, &tp.CreatedAt, &tp.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get teacher partner by id: %w", err)
	}
	return &tp, nil
}

func (r *repository) ListTeacherPartners(ctx context.Context, query TeacherPartnerQuery) (*TeacherPartnerPageResult, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	where, args := buildTeacherPartnerWhere(query)

	var total int64
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM teacher_partners "+where, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count teacher partners: %w", err)
	}

	queryArgs := append([]any{}, args...)
	limitIndex := len(queryArgs) + 1
	offsetIndex := len(queryArgs) + 2
	queryArgs = append(queryArgs, query.PageSize, (query.PageNum-1)*query.PageSize)

	rows, err := exec.QueryContext(ctx,
		fmt.Sprintf(`SELECT tp.id, tp.teacher_id, tp.partner_id, tp.partner_config_id, tp.cooperation_type, tp.commission_rate, tp.start_date, tp.end_date, tp.status, tp.remark, tp.created_at, pc.name
		 FROM teacher_partners tp LEFT JOIN partner_configs pc ON tp.partner_id = pc.id %s ORDER BY tp.created_at DESC LIMIT $%d OFFSET $%d`, where, limitIndex, offsetIndex),
		queryArgs...,
	)
	if err != nil {
		return nil, fmt.Errorf("list teacher partners: %w", err)
	}
	defer rows.Close()

	records := []TeacherPartnerVO{}
	for rows.Next() {
		var tp TeacherPartner
		var partnerName string
		if err := rows.Scan(&tp.ID, &tp.TeacherID, &tp.PartnerID, &tp.PartnerConfigID, &tp.CooperationType, &tp.CommissionRate,
			&tp.StartDate, &tp.EndDate, &tp.Status, &tp.Remark, &tp.CreatedAt, &partnerName); err != nil {
			return nil, fmt.Errorf("scan teacher partner: %w", err)
		}
		records = append(records, toTeacherPartnerVO(tp, partnerName))
	}
	return &TeacherPartnerPageResult{Total: total, PageNum: query.PageNum, PageSize: query.PageSize, Records: records}, nil
}

func (r *repository) UpdateTeacherPartner(ctx context.Context, id int64, updates map[string]any) error {
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
	query := fmt.Sprintf("UPDATE teacher_partners SET %s, updated_at = NOW() WHERE id = $%d", strings.Join(sets, ", "), len(args))
	_, err := exec.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update teacher partner: %w", err)
	}
	return nil
}

func (r *repository) DeleteTeacherPartner(ctx context.Context, id int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx, `DELETE FROM teacher_partners WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete teacher partner: %w", err)
	}
	return nil
}

func (r *repository) ListPartneredTeachers(ctx context.Context, page, pageSize int) (*TeacherPartnerPageResult, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	var total int64
	if err := exec.QueryRowContext(ctx, `SELECT COUNT(DISTINCT teacher_id) FROM teacher_partners WHERE status = $1`, CooperationStatusActive).Scan(&total); err != nil {
		return nil, fmt.Errorf("count partnered teachers: %w", err)
	}

	rows, err := exec.QueryContext(ctx,
		`SELECT tp.id, tp.teacher_id, tp.partner_id, tp.partner_config_id, tp.cooperation_type, tp.commission_rate, tp.start_date, tp.end_date, tp.status, tp.remark, tp.created_at, pc.name
		 FROM teacher_partners tp LEFT JOIN partner_configs pc ON tp.partner_id = pc.id
		 WHERE tp.status = $1 ORDER BY tp.created_at DESC LIMIT $2 OFFSET $3`,
		CooperationStatusActive, pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, fmt.Errorf("list partnered teachers: %w", err)
	}
	defer rows.Close()

	records := []TeacherPartnerVO{}
	for rows.Next() {
		var tp TeacherPartner
		var partnerName string
		if err := rows.Scan(&tp.ID, &tp.TeacherID, &tp.PartnerID, &tp.PartnerConfigID, &tp.CooperationType, &tp.CommissionRate,
			&tp.StartDate, &tp.EndDate, &tp.Status, &tp.Remark, &tp.CreatedAt, &partnerName); err != nil {
			return nil, fmt.Errorf("scan partnered teacher: %w", err)
		}
		records = append(records, toTeacherPartnerVO(tp, partnerName))
	}
	return &TeacherPartnerPageResult{Total: total, PageNum: page, PageSize: pageSize, Records: records}, nil
}

func buildPartnerConfigWhere(query PartnerConfigQuery) (string, []any) {
	where := "WHERE 1=1"
	args := []any{}
	if query.Name != "" {
		args = append(args, "%"+query.Name+"%")
		where += fmt.Sprintf(" AND name ILIKE $%d", len(args))
	}
	if query.Status != "" {
		args = append(args, query.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	if query.PartnerType != "" {
		args = append(args, query.PartnerType)
		where += fmt.Sprintf(" AND partner_type = $%d", len(args))
	}
	return where, args
}

func buildTeacherPartnerWhere(query TeacherPartnerQuery) (string, []any) {
	where := "WHERE 1=1"
	args := []any{}
	if query.TeacherID != 0 {
		args = append(args, query.TeacherID)
		where += fmt.Sprintf(" AND teacher_id = $%d", len(args))
	}
	if query.PartnerID != 0 {
		args = append(args, query.PartnerID)
		where += fmt.Sprintf(" AND partner_id = $%d", len(args))
	}
	if query.Status != "" {
		args = append(args, query.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}
	return where, args
}
