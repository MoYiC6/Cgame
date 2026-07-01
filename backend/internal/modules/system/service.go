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

// RealName Verify Log
func (s *Service) CreateRealNameVerifyLog(ctx context.Context, log *RealNameVerifyLog) error {
	return s.repo.CreateRealNameVerifyLog(ctx, log)
}

func (s *Service) ListRealNameVerifyLogs(ctx context.Context, userID *int64, eventType string, page, pageSize int) ([]*RealNameVerifyLog, int, error) {
	return s.repo.ListRealNameVerifyLogs(ctx, userID, eventType, page, pageSize)
}

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
