package system

import (
	"context"
	"fmt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// System Settings
func (s *Service) GetSetting(ctx context.Context, key string) (string, error) {
	setting, err := s.repo.GetSetting(ctx, key)
	if err != nil {
		return "", nil
	}
	if setting.Value != nil {
		return *setting.Value, nil
	}
	return "", nil
}

func (s *Service) SetSetting(ctx context.Context, key, value string) error {
	setting := &SystemSetting{
		Key:   key,
		Value: &value,
		Type:  strPtr("string"),
	}
	return s.repo.SetSetting(ctx, setting)
}

func (s *Service) ListSettings(ctx context.Context, prefix string) ([]*SystemSetting, error) {
	return s.repo.ListSettings(ctx, prefix)
}

func (s *Service) GetAllSettings(ctx context.Context) ([]*SystemSetting, error) {
	return s.repo.GetAllSettings(ctx)
}

// Partner Config
func (s *Service) GetPartnerConfig(ctx context.Context, key string) (string, error) {
	config, err := s.repo.GetPartnerConfig(ctx, key)
	if err != nil {
		return "", nil
	}
	if config.Value != nil {
		return *config.Value, nil
	}
	return "", nil
}

func (s *Service) SetPartnerConfig(ctx context.Context, key, value string) error {
	config := &PartnerConfig{
		Key:    key,
		Value:  &value,
		Status: intPtr(1),
	}
	return s.repo.SetPartnerConfig(ctx, config)
}

// FaceId Config
func (s *Service) GetActiveFaceIdConfig(ctx context.Context) (*FaceIdConfig, error) {
	config, err := s.repo.GetActiveFaceIdConfig(ctx)
	if err != nil {
		return nil, nil
	}
	return config, nil
}

func (s *Service) CreateFaceIdConfig(ctx context.Context, c *FaceIdConfig) (int64, error) {
	if c.SecretID == "" || c.SecretKey == "" || c.RuleID == "" {
		return 0, fmt.Errorf("secret_id, secret_key and rule_id are required")
	}
	if c.Region == nil || *c.Region == "" {
		c.Region = strPtr("ap-guangzhou")
	}
	if c.ManualEnabled == nil {
		c.ManualEnabled = intPtr(1)
	}
	if err := s.repo.CreateFaceIdConfig(ctx, c); err != nil {
		return 0, err
	}
	return c.ID, nil
}

func (s *Service) UpdateFaceIdConfig(ctx context.Context, c *FaceIdConfig) error {
	if c.ID == 0 {
		return fmt.Errorf("id is required")
	}
	return s.repo.UpdateFaceIdConfig(ctx, c)
}

func (s *Service) DeleteFaceIdConfig(ctx context.Context, id int64) error {
	return s.repo.DeleteFaceIdConfig(ctx, id)
}

func (s *Service) ListFaceIdConfigs(ctx context.Context, page, pageSize int) ([]*FaceIdConfig, int, error) {
	return s.repo.ListFaceIdConfigs(ctx, page, pageSize)
}

// Customer Service Config
func (s *Service) GetCustomerServiceConfig(ctx context.Context) (map[string]any, error) {
	configType, _ := s.GetSetting(ctx, "customer_service.type")
	if configType == "" {
		configType = "fallback"
	}
	enabledStr, _ := s.GetSetting(ctx, "customer_service.enabled")
	enabled := enabledStr != "false"
	phone, _ := s.GetSetting(ctx, "customer_service.phone")
	wechat, _ := s.GetSetting(ctx, "customer_service.wechat")
	qq, _ := s.GetSetting(ctx, "customer_service.qq")
	email, _ := s.GetSetting(ctx, "customer_service.email")
	workHours, _ := s.GetSetting(ctx, "customer_service.work_hours")
	if workHours == "" {
		workHours = "9:00-22:00"
	}
	return map[string]any{
		"type":      configType,
		"enabled":   enabled,
		"phone":     phone,
		"wechat":    wechat,
		"qq":        qq,
		"email":     email,
		"workHours": workHours,
	}, nil
}

// RealName Verify Log
func (s *Service) CreateRealNameVerifyLog(ctx context.Context, log *RealNameVerifyLog) error {
	return s.repo.CreateRealNameVerifyLog(ctx, log)
}

func (s *Service) ListRealNameVerifyLogs(ctx context.Context, userID *int64, eventType string, page, pageSize int) ([]*RealNameVerifyLog, int, error) {
	return s.repo.ListRealNameVerifyLogs(ctx, userID, eventType, page, pageSize)
}

// Menu service methods
func (s *Service) GetMenuTree(ctx context.Context) ([]*SystemMenu, error) {
	menus, err := s.repo.ListMenus(ctx)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(menus), nil
}

func (s *Service) ListMenus(ctx context.Context) ([]*SystemMenu, error) {
	return s.repo.ListMenus(ctx)
}

func (s *Service) GetMenuByID(ctx context.Context, id int64) (*SystemMenu, error) {
	return s.repo.GetMenuByID(ctx, id)
}

func (s *Service) BatchCreateMenus(ctx context.Context, menus []SystemMenu) error {
	for i := range menus {
		if err := s.repo.CreateMenu(ctx, &menus[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) UpdateMenu(ctx context.Context, m *SystemMenu) error {
	if m.ID == 0 {
		return fmt.Errorf("id is required")
	}
	return s.repo.UpdateMenu(ctx, m)
}

func (s *Service) DeleteMenu(ctx context.Context, id int64) error {
	return s.repo.DeleteMenu(ctx, id)
}

func (s *Service) GetMenuCascader(ctx context.Context) ([]*SystemMenu, error) {
	menus, err := s.repo.ListMenus(ctx)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(menus), nil
}

func buildMenuTree(menus []*SystemMenu) []*SystemMenu {
	menuMap := make(map[int64]*SystemMenu)
	var roots []*SystemMenu
	for _, m := range menus {
		menuMap[m.ID] = m
	}
	for _, m := range menus {
		if m.ParentID != nil && *m.ParentID != 0 {
			if parent, ok := menuMap[*m.ParentID]; ok {
				if parent.Children == nil {
					parent.Children = []*SystemMenu{}
				}
				parent.Children = append(parent.Children, m)
			}
		} else {
			roots = append(roots, m)
		}
	}
	return roots
}

// Permission service methods
func (s *Service) ListPermissions(ctx context.Context) ([]*SystemPermission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *Service) UpdatePermission(ctx context.Context, p *SystemPermission) error {
	if p.ID == 0 {
		return fmt.Errorf("id is required")
	}
	return s.repo.UpdatePermission(ctx, p)
}

func (s *Service) DeletePermission(ctx context.Context, id int64) error {
	return s.repo.DeletePermission(ctx, id)
}

// Role service methods
func (s *Service) ListRoles(ctx context.Context) ([]*SystemRole, error) {
	return s.repo.ListRoles(ctx)
}

func (s *Service) CreateRole(ctx context.Context, role *SystemRole) (int64, error) {
	if role.Name == "" || role.Code == "" {
		return 0, fmt.Errorf("name and code are required")
	}
	if err := s.repo.CreateRole(ctx, role); err != nil {
		return 0, err
	}
	return role.ID, nil
}

func (s *Service) UpdateRole(ctx context.Context, role *SystemRole) error {
	if role.ID == 0 {
		return fmt.Errorf("id is required")
	}
	return s.repo.UpdateRole(ctx, role)
}

func (s *Service) DeleteRole(ctx context.Context, id int64) error {
	return s.repo.DeleteRole(ctx, id)
}

func (s *Service) UpdateRoleStatus(ctx context.Context, id int64, status int) error {
	return s.repo.UpdateRoleStatus(ctx, id, status)
}

func (s *Service) AssignPermissionsToRole(ctx context.Context, roleID int64, permissionIDs []int64) error {
	return s.repo.AssignPermissionsToRole(ctx, roleID, permissionIDs)
}

func (s *Service) AssignMenusToRole(ctx context.Context, roleID int64, menuIDs []int64) error {
	return s.repo.AssignMenusToRole(ctx, roleID, menuIDs)
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func (s *Service) UpdateFaceIdStatus(ctx context.Context, id int64, enabled int) error {
	if id == 0 {
		return fmt.Errorf("id is required")
	}
	return s.repo.UpdateFaceIdStatus(ctx, id, enabled)
}
