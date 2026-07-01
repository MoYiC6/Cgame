package system

import (
	"context"
	"fmt"

	"backend/internal/platform/database"
)

type Repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) *Repository {
	return &Repository{dbtx: dbtx}
}

// System Settings
func (r *Repository) GetSetting(ctx context.Context, key string) (*SystemSetting, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, setting_key, setting_value, setting_type, category, description, is_public, created_at, updated_at
		 FROM system_settings WHERE setting_key = $1`,
		key,
	)
	var s SystemSetting
	err := row.Scan(&s.ID, &s.Key, &s.Value, &s.Type, &s.Category, &s.Description, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get setting: %w", err)
	}
	return &s, nil
}

func (r *Repository) SetSetting(ctx context.Context, s *SystemSetting) error {
	_, err := r.dbtx.ExecContext(ctx,
		`INSERT INTO system_settings (setting_key, setting_value, setting_type, category, description, is_public)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (setting_key) DO UPDATE SET
			setting_value = EXCLUDED.setting_value,
			updated_at = NOW()`,
		s.Key, s.Value, s.Type, s.Category, s.Description, s.IsPublic,
	)
	if err != nil {
		return fmt.Errorf("set setting: %w", err)
	}
	return nil
}

func (r *Repository) ListSettings(ctx context.Context, prefix string) ([]*SystemSetting, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, setting_key, setting_value, setting_type, category, description, is_public, created_at, updated_at
		 FROM system_settings WHERE setting_key LIKE $1 || '%' ORDER BY id ASC`,
		prefix,
	)
	if err != nil {
		return nil, fmt.Errorf("list settings: %w", err)
	}
	defer rows.Close()

	var settings []*SystemSetting
	for rows.Next() {
		var s SystemSetting
		if err := rows.Scan(&s.ID, &s.Key, &s.Value, &s.Type, &s.Category, &s.Description, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		settings = append(settings, &s)
	}
	return settings, nil
}

func (r *Repository) GetAllSettings(ctx context.Context) ([]*SystemSetting, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, setting_key, setting_value, setting_type, category, description, is_public, created_at, updated_at FROM system_settings ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("get all settings: %w", err)
	}
	defer rows.Close()

	var settings []*SystemSetting
	for rows.Next() {
		var s SystemSetting
		if err := rows.Scan(&s.ID, &s.Key, &s.Value, &s.Type, &s.Category, &s.Description, &s.IsPublic, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		settings = append(settings, &s)
	}
	return settings, nil
}

// Partner Config
func (r *Repository) GetPartnerConfig(ctx context.Context, key string) (*PartnerConfig, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, config_key, config_value, description, status, created_at, updated_at FROM partner_config WHERE config_key = $1`,
		key,
	)
	var c PartnerConfig
	err := row.Scan(&c.ID, &c.Key, &c.Value, &c.Description, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get partner config: %w", err)
	}
	return &c, nil
}

func (r *Repository) SetPartnerConfig(ctx context.Context, c *PartnerConfig) error {
	_, err := r.dbtx.ExecContext(ctx,
		`INSERT INTO partner_config (config_key, config_value, description, status)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (config_key) DO UPDATE SET
			config_value = EXCLUDED.config_value,
			updated_at = NOW()`,
		c.Key, c.Value, c.Description, c.Status,
	)
	if err != nil {
		return fmt.Errorf("set partner config: %w", err)
	}
	return nil
}

// FaceId Config
func (r *Repository) GetActiveFaceIdConfig(ctx context.Context) (*FaceIdConfig, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, secret_id, secret_key, rule_id, region, redirect_url, is_enabled, remark, manual_enabled, created_at, updated_at, deleted
		 FROM faceid_config WHERE is_enabled = 1 AND deleted = 0 ORDER BY updated_at DESC LIMIT 1`,
	)
	var c FaceIdConfig
	err := row.Scan(&c.ID, &c.SecretID, &c.SecretKey, &c.RuleID, &c.Region, &c.RedirectURL, &c.IsEnabled, &c.Remark, &c.ManualEnabled, &c.CreatedAt, &c.UpdatedAt, &c.Deleted)
	if err != nil {
		return nil, fmt.Errorf("get active faceid config: %w", err)
	}
	return &c, nil
}

func (r *Repository) CreateFaceIdConfig(ctx context.Context, c *FaceIdConfig) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO faceid_config (secret_id, secret_key, rule_id, region, redirect_url, is_enabled, remark, manual_enabled)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		c.SecretID, c.SecretKey, c.RuleID, c.Region, c.RedirectURL, c.IsEnabled, c.Remark, c.ManualEnabled,
	).Scan(&c.ID)
}

func (r *Repository) UpdateFaceIdConfig(ctx context.Context, c *FaceIdConfig) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE faceid_config SET secret_id = $1, secret_key = $2, rule_id = $3, region = $4, redirect_url = $5,
		 is_enabled = $6, remark = $7, manual_enabled = $8, updated_at = NOW() WHERE id = $9 AND deleted = 0`,
		c.SecretID, c.SecretKey, c.RuleID, c.Region, c.RedirectURL, c.IsEnabled, c.Remark, c.ManualEnabled, c.ID,
	)
	if err != nil {
		return fmt.Errorf("update faceid config: %w", err)
	}
	return nil
}

func (r *Repository) DeleteFaceIdConfig(ctx context.Context, id int64) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE faceid_config SET deleted = 1, updated_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete faceid config: %w", err)
	}
	return nil
}

// RealName Verify Log
func (r *Repository) CreateRealNameVerifyLog(ctx context.Context, log *RealNameVerifyLog) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO realname_verify_log (user_id, event_type, operator_id, operator_type, detail, ip_address, submitted_name, submitted_id_card)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		log.UserID, log.EventType, log.OperatorID, log.OperatorType, log.Detail, log.IPAddress, log.SubmittedName, log.SubmittedIDCard,
	).Scan(&log.ID)
}

func (r *Repository) ListRealNameVerifyLogs(ctx context.Context, userID *int64, eventType string, page, pageSize int) ([]*RealNameVerifyLog, int, error) {
	where := "WHERE deleted = 0"
	args := []interface{}{}
	idx := 1

	if userID != nil {
		where += fmt.Sprintf(" AND user_id = $%d", idx)
		args = append(args, *userID)
		idx++
	}
	if eventType != "" {
		where += fmt.Sprintf(" AND event_type = $%d", idx)
		args = append(args, eventType)
		idx++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM realname_verify_log %s", where)
	var total int
	if err := r.dbtx.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count logs: %w", err)
	}

	query := fmt.Sprintf(
		`SELECT id, user_id, event_type, operator_id, operator_type, detail, ip_address, submitted_name, submitted_id_card, created_at, updated_at, deleted
		 FROM realname_verify_log %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1,
	)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := r.dbtx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list logs: %w", err)
	}
	defer rows.Close()

	var logs []*RealNameVerifyLog
	for rows.Next() {
		var l RealNameVerifyLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.EventType, &l.OperatorID, &l.OperatorType, &l.Detail, &l.IPAddress, &l.SubmittedName, &l.SubmittedIDCard, &l.CreatedAt, &l.UpdatedAt, &l.Deleted); err != nil {
			return nil, 0, fmt.Errorf("scan log: %w", err)
		}
		logs = append(logs, &l)
	}
	return logs, total, nil
}
