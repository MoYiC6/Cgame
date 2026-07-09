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

func (r *Repository) ListFaceIdConfigs(ctx context.Context, page, pageSize int) ([]*FaceIdConfig, int, error) {
	var total int
	if err := r.dbtx.QueryRowContext(ctx, "SELECT COUNT(*) FROM faceid_config WHERE deleted = 0").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count faceid configs: %w", err)
	}

	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, secret_id, secret_key, rule_id, region, redirect_url, is_enabled, remark, manual_enabled, created_at, updated_at, deleted
		 FROM faceid_config WHERE deleted = 0 ORDER BY updated_at DESC LIMIT $1 OFFSET $2`,
		pageSize, (page-1)*pageSize,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list faceid configs: %w", err)
	}
	defer rows.Close()

	var configs []*FaceIdConfig
	for rows.Next() {
		var c FaceIdConfig
		if err := rows.Scan(&c.ID, &c.SecretID, &c.SecretKey, &c.RuleID, &c.Region, &c.RedirectURL, &c.IsEnabled, &c.Remark, &c.ManualEnabled, &c.CreatedAt, &c.UpdatedAt, &c.Deleted); err != nil {
			return nil, 0, fmt.Errorf("scan faceid config: %w", err)
		}
		configs = append(configs, &c)
	}
	return configs, total, nil
}

func (r *Repository) UpdateFaceIdStatus(ctx context.Context, id int64, enabled int) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE faceid_config SET is_enabled = $1, updated_at = NOW() WHERE id = $2 AND deleted = 0`,
		enabled, id,
	)
	if err != nil {
		return fmt.Errorf("update faceid status: %w", err)
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

// Menu CRUD
func (r *Repository) CreateMenu(ctx context.Context, m *SystemMenu) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO system_menus (parent_id, name, path, component, icon, sort, status, menu_type, permission_code)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		m.ParentID, m.Name, m.Path, m.Component, m.Icon, m.Sort, m.Status, m.MenuType, m.PermissionCode,
	).Scan(&m.ID)
}

func (r *Repository) GetMenuByID(ctx context.Context, id int64) (*SystemMenu, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, parent_id, name, path, component, icon, sort, status, menu_type, permission_code, created_at, updated_at
		 FROM system_menus WHERE id = $1`,
		id,
	)
	var m SystemMenu
	err := row.Scan(&m.ID, &m.ParentID, &m.Name, &m.Path, &m.Component, &m.Icon, &m.Sort, &m.Status, &m.MenuType, &m.PermissionCode, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get menu: %w", err)
	}
	return &m, nil
}

func (r *Repository) UpdateMenu(ctx context.Context, m *SystemMenu) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE system_menus SET parent_id = $1, name = $2, path = $3, component = $4, icon = $5, sort = $6, status = $7, menu_type = $8, permission_code = $9, updated_at = NOW() WHERE id = $10`,
		m.ParentID, m.Name, m.Path, m.Component, m.Icon, m.Sort, m.Status, m.MenuType, m.PermissionCode, m.ID,
	)
	if err != nil {
		return fmt.Errorf("update menu: %w", err)
	}
	return nil
}

func (r *Repository) DeleteMenu(ctx context.Context, id int64) error {
	_, err := r.dbtx.ExecContext(ctx, "DELETE FROM system_menus WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete menu: %w", err)
	}
	return nil
}

func (r *Repository) ListMenus(ctx context.Context) ([]*SystemMenu, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, parent_id, name, path, component, icon, sort, status, menu_type, permission_code, created_at, updated_at
		 FROM system_menus ORDER BY sort ASC, id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list menus: %w", err)
	}
	defer rows.Close()

	var menus []*SystemMenu
	for rows.Next() {
		var m SystemMenu
		if err := rows.Scan(&m.ID, &m.ParentID, &m.Name, &m.Path, &m.Component, &m.Icon, &m.Sort, &m.Status, &m.MenuType, &m.PermissionCode, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan menu: %w", err)
		}
		menus = append(menus, &m)
	}
	return menus, nil
}

// Permission CRUD
func (r *Repository) CreatePermission(ctx context.Context, p *SystemPermission) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO system_permissions (name, code, description, status)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		p.Name, p.Code, p.Description, p.Status,
	).Scan(&p.ID)
}

func (r *Repository) GetPermissionByID(ctx context.Context, id int64) (*SystemPermission, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, name, code, description, status, created_at, updated_at
		 FROM system_permissions WHERE id = $1`,
		id,
	)
	var p SystemPermission
	err := row.Scan(&p.ID, &p.Name, &p.Code, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get permission: %w", err)
	}
	return &p, nil
}

func (r *Repository) UpdatePermission(ctx context.Context, p *SystemPermission) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE system_permissions SET name = $1, code = $2, description = $3, status = $4, updated_at = NOW() WHERE id = $5`,
		p.Name, p.Code, p.Description, p.Status, p.ID,
	)
	if err != nil {
		return fmt.Errorf("update permission: %w", err)
	}
	return nil
}

func (r *Repository) DeletePermission(ctx context.Context, id int64) error {
	_, err := r.dbtx.ExecContext(ctx, "DELETE FROM system_permissions WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete permission: %w", err)
	}
	return nil
}

func (r *Repository) ListPermissions(ctx context.Context) ([]*SystemPermission, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, name, code, description, status, created_at, updated_at
		 FROM system_permissions ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*SystemPermission
	for rows.Next() {
		var p SystemPermission
		if err := rows.Scan(&p.ID, &p.Name, &p.Code, &p.Description, &p.Status, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		permissions = append(permissions, &p)
	}
	return permissions, nil
}

// Role CRUD
func (r *Repository) CreateRole(ctx context.Context, role *SystemRole) error {
	return r.dbtx.QueryRowContext(ctx,
		`INSERT INTO system_roles (name, code, description, status)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		role.Name, role.Code, role.Description, role.Status,
	).Scan(&role.ID)
}

func (r *Repository) GetRoleByID(ctx context.Context, id int64) (*SystemRole, error) {
	row := r.dbtx.QueryRowContext(ctx,
		`SELECT id, name, code, description, status, created_at, updated_at
		 FROM system_roles WHERE id = $1`,
		id,
	)
	var role SystemRole
	err := row.Scan(&role.ID, &role.Name, &role.Code, &role.Description, &role.Status, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}
	return &role, nil
}

func (r *Repository) UpdateRole(ctx context.Context, role *SystemRole) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE system_roles SET name = $1, code = $2, description = $3, status = $4, updated_at = NOW() WHERE id = $5`,
		role.Name, role.Code, role.Description, role.Status, role.ID,
	)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	return nil
}

func (r *Repository) DeleteRole(ctx context.Context, id int64) error {
	_, err := r.dbtx.ExecContext(ctx, "DELETE FROM system_roles WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}

func (r *Repository) ListRoles(ctx context.Context) ([]*SystemRole, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT id, name, code, description, status, created_at, updated_at
		 FROM system_roles ORDER BY id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()

	var roles []*SystemRole
	for rows.Next() {
		var role SystemRole
		if err := rows.Scan(&role.ID, &role.Name, &role.Code, &role.Description, &role.Status, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, &role)
	}
	return roles, nil
}

func (r *Repository) UpdateRoleStatus(ctx context.Context, id int64, status int) error {
	_, err := r.dbtx.ExecContext(ctx,
		`UPDATE system_roles SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update role status: %w", err)
	}
	return nil
}

// Role associations
func (r *Repository) AssignPermissionsToRole(ctx context.Context, roleID int64, permissionIDs []int64) error {
	_, err := r.dbtx.ExecContext(ctx, "DELETE FROM role_permissions WHERE role_id = $1", roleID)
	if err != nil {
		return fmt.Errorf("clear role permissions: %w", err)
	}
	for _, pid := range permissionIDs {
		_, err := r.dbtx.ExecContext(ctx,
			`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`,
			roleID, pid,
		)
		if err != nil {
			return fmt.Errorf("assign permission: %w", err)
		}
	}
	return nil
}

func (r *Repository) AssignMenusToRole(ctx context.Context, roleID int64, menuIDs []int64) error {
	_, err := r.dbtx.ExecContext(ctx, "DELETE FROM role_menus WHERE role_id = $1", roleID)
	if err != nil {
		return fmt.Errorf("clear role menus: %w", err)
	}
	for _, mid := range menuIDs {
		_, err := r.dbtx.ExecContext(ctx,
			`INSERT INTO role_menus (role_id, menu_id) VALUES ($1, $2)`,
			roleID, mid,
		)
		if err != nil {
			return fmt.Errorf("assign menu: %w", err)
		}
	}
	return nil
}

func (r *Repository) GetRolePermissions(ctx context.Context, roleID int64) ([]int64, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT permission_id FROM role_permissions WHERE role_id = $1`,
		roleID,
	)
	if err != nil {
		return nil, fmt.Errorf("get role permissions: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan permission id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *Repository) GetRoleMenus(ctx context.Context, roleID int64) ([]int64, error) {
	rows, err := r.dbtx.QueryContext(ctx,
		`SELECT menu_id FROM role_menus WHERE role_id = $1`,
		roleID,
	)
	if err != nil {
		return nil, fmt.Errorf("get role menus: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan menu id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
