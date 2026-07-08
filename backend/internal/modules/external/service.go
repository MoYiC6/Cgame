package external

import (
	"context"
	"fmt"
	"time"
)

const (
	PlatformWechat = "wechat"
	PlatformKook   = "kook"
)

type WxPayConfig struct {
	ID         int64
	ConfigType string
	AppID      string
	MchID      string
	APIKey     string
	APIV3Key   string
	SerialNo   string
	PrivateKey string
	PublicKey  string
	Status     int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Repository interface {
	GetUserOAuth(ctx context.Context, platform, openID string) (*UserOAuth, error)
	GetUserOAuthByUserID(ctx context.Context, userID int64, platform string) (*UserOAuth, error)
	CreateUserOAuth(ctx context.Context, oauth *UserOAuth) error
	UpdateUserOAuth(ctx context.Context, oauth *UserOAuth) error
	DeleteUserOAuth(ctx context.Context, userID int64, platform string) error
	CreateUserToken(ctx context.Context, token *UserToken) error
	GetUserToken(ctx context.Context, accessToken string) (*UserToken, error)
	CreateScanLoginSession(ctx context.Context, session *ScanLoginSession) error
	GetScanLoginSession(ctx context.Context, loginKey string) (*ScanLoginSession, error)
	UpdateScanLoginSession(ctx context.Context, session *ScanLoginSession) error
	CreateWxPayConfig(ctx context.Context, config *WxPayConfig) error
	GetWxPayConfig(ctx context.Context, configType string) (*WxPayConfig, error)
	GetWxPayConfigByID(ctx context.Context, id int64) (*WxPayConfig, error)
	ListWxPayConfigs(ctx context.Context, page, pageSize int, configType *string) ([]*WxPayConfig, int, error)
	UpdateWxPayConfig(ctx context.Context, config *WxPayConfig) error
	UpdateWxPayConfigStatus(ctx context.Context, id int64, status int) error
	DeleteWxPayConfig(ctx context.Context, id int64) error
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) WechatLogin(ctx context.Context, platform, code, appID string) (*UserOAuth, *UserToken, error) {
	return nil, nil, fmt.Errorf("not implemented: wechat login requires external API call")
}

func (s *Service) WechatBind(ctx context.Context, userID int64, platform, code string) (*UserOAuth, error) {
	return nil, fmt.Errorf("not implemented: wechat bind requires external API call")
}

func (s *Service) WechatUnbind(ctx context.Context, userID int64, platform string) error {
	return s.repo.DeleteUserOAuth(ctx, userID, platform)
}

func (s *Service) GetWechatPhone(ctx context.Context, code string) (*WechatPhoneResponse, error) {
	return nil, fmt.Errorf("not implemented: wechat phone requires external API call")
}

func (s *Service) GenerateScanLoginQR(ctx context.Context) (*ScanLoginSession, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Service) CheckScanLoginStatus(ctx context.Context, loginKey string) (*ScanLoginSession, error) {
	return s.repo.GetScanLoginSession(ctx, loginKey)
}

func (s *Service) ConfirmScanLogin(ctx context.Context, loginKey string, userID int64, token string) error {
	session, err := s.repo.GetScanLoginSession(ctx, loginKey)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	session.Status = "success"
	session.UserID = &userID
	session.Token = &token
	return s.repo.UpdateScanLoginSession(ctx, session)
}

func (s *Service) CreateWxPayConfig(ctx context.Context, config *WxPayConfig) (int64, error) {
	if config.ConfigType == "" || config.AppID == "" {
		return 0, fmt.Errorf("config_type and app_id are required")
	}
	if err := s.repo.CreateWxPayConfig(ctx, config); err != nil {
		return 0, fmt.Errorf("create wx pay config: %w", err)
	}
	return config.ID, nil
}

func (s *Service) GetWxPayConfig(ctx context.Context, configType string) (*WxPayConfig, error) {
	config, err := s.repo.GetWxPayConfig(ctx, configType)
	if err != nil {
		return nil, fmt.Errorf("get wx pay config: %w", err)
	}
	return config, nil
}

func (s *Service) GetWxPayConfigByID(ctx context.Context, id int64) (*WxPayConfig, error) {
	if id == 0 {
		return nil, fmt.Errorf("id is required")
	}
	config, err := s.repo.GetWxPayConfigByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get wx pay config by id: %w", err)
	}
	return config, nil
}

func (s *Service) ListWxPayConfigs(ctx context.Context, page, pageSize int, configType *string) ([]*WxPayConfig, int, error) {
	configs, total, err := s.repo.ListWxPayConfigs(ctx, page, pageSize, configType)
	if err != nil {
		return nil, 0, fmt.Errorf("list wx pay configs: %w", err)
	}
	return configs, total, nil
}

func (s *Service) UpdateWxPayConfig(ctx context.Context, config *WxPayConfig) error {
	if config.ID == 0 {
		return fmt.Errorf("id is required")
	}
	if err := s.repo.UpdateWxPayConfig(ctx, config); err != nil {
		return fmt.Errorf("update wx pay config: %w", err)
	}
	return nil
}

func (s *Service) UpdateWxPayConfigStatus(ctx context.Context, id int64, status int) error {
	if id == 0 {
		return fmt.Errorf("id is required")
	}
	if status != 0 && status != 1 {
		return fmt.Errorf("status must be 0 or 1")
	}
	if err := s.repo.UpdateWxPayConfigStatus(ctx, id, status); err != nil {
		return fmt.Errorf("update wx pay config status: %w", err)
	}
	return nil
}

func (s *Service) DeleteWxPayConfig(ctx context.Context, id int64) error {
	if err := s.repo.DeleteWxPayConfig(ctx, id); err != nil {
		return fmt.Errorf("delete wx pay config: %w", err)
	}
	return nil
}
