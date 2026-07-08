package external

import (
	"context"
	"fmt"

	"backend/internal/platform/database"
)

type repository struct {
	dbtx database.DBTX
}

func NewRepository(dbtx database.DBTX) Repository {
	return &repository{dbtx: dbtx}
}

func (r *repository) GetUserOAuth(ctx context.Context, platform, openID string) (*UserOAuth, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, user_id, platform, open_id, union_id, nickname, avatar, session_key, phone, bound_at, created_at, updated_at
		 FROM user_oauth WHERE platform = $1 AND open_id = $2`,
		platform, openID,
	)
	var oauth UserOAuth
	err := row.Scan(&oauth.ID, &oauth.UserID, &oauth.Platform, &oauth.OpenID, &oauth.UnionID, &oauth.Nickname, &oauth.Avatar, &oauth.SessionKey, &oauth.Phone, &oauth.BoundAt, &oauth.CreatedAt, &oauth.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user oauth: %w", err)
	}
	return &oauth, nil
}

func (r *repository) GetUserOAuthByUserID(ctx context.Context, userID int64, platform string) (*UserOAuth, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, user_id, platform, open_id, union_id, nickname, avatar, session_key, phone, bound_at, created_at, updated_at
		 FROM user_oauth WHERE user_id = $1 AND platform = $2`,
		userID, platform,
	)
	var oauth UserOAuth
	err := row.Scan(&oauth.ID, &oauth.UserID, &oauth.Platform, &oauth.OpenID, &oauth.UnionID, &oauth.Nickname, &oauth.Avatar, &oauth.SessionKey, &oauth.Phone, &oauth.BoundAt, &oauth.CreatedAt, &oauth.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user oauth by user id: %w", err)
	}
	return &oauth, nil
}

func (r *repository) CreateUserOAuth(ctx context.Context, oauth *UserOAuth) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO user_oauth (user_id, platform, open_id, union_id, nickname, avatar, session_key, phone, bound_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		oauth.UserID, oauth.Platform, oauth.OpenID, oauth.UnionID, oauth.Nickname, oauth.Avatar, oauth.SessionKey, oauth.Phone, oauth.BoundAt,
	).Scan(&oauth.ID)
}

func (r *repository) UpdateUserOAuth(ctx context.Context, oauth *UserOAuth) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE user_oauth SET union_id = $1, nickname = $2, avatar = $3, session_key = $4, phone = $5, updated_at = NOW() WHERE id = $6`,
		oauth.UnionID, oauth.Nickname, oauth.Avatar, oauth.SessionKey, oauth.Phone, oauth.ID,
	)
	if err != nil {
		return fmt.Errorf("update user oauth: %w", err)
	}
	return nil
}

func (r *repository) DeleteUserOAuth(ctx context.Context, userID int64, platform string) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`DELETE FROM user_oauth WHERE user_id = $1 AND platform = $2`,
		userID, platform,
	)
	if err != nil {
		return fmt.Errorf("delete user oauth: %w", err)
	}
	return nil
}

func (r *repository) CreateUserToken(ctx context.Context, token *UserToken) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO user_tokens (user_id, access_token, refresh_token, expires_at) VALUES ($1, $2, $3, $4) RETURNING id`,
		token.UserID, token.AccessToken, token.RefreshToken, token.ExpiresAt,
	).Scan(&token.ID)
}

func (r *repository) GetUserToken(ctx context.Context, accessToken string) (*UserToken, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, user_id, access_token, refresh_token, expires_at, created_at FROM user_tokens WHERE access_token = $1`,
		accessToken,
	)
	var token UserToken
	err := row.Scan(&token.ID, &token.UserID, &token.AccessToken, &token.RefreshToken, &token.ExpiresAt, &token.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get user token: %w", err)
	}
	return &token, nil
}

func (r *repository) CreateScanLoginSession(ctx context.Context, session *ScanLoginSession) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`INSERT INTO scan_login_sessions (login_key, status, expires_at) VALUES ($1, $2, $3)`,
		session.LoginKey, session.Status, session.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("create scan login session: %w", err)
	}
	return nil
}

func (r *repository) GetScanLoginSession(ctx context.Context, loginKey string) (*ScanLoginSession, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT login_key, status, user_id, token, expires_at, created_at, updated_at FROM scan_login_sessions WHERE login_key = $1`,
		loginKey,
	)
	var session ScanLoginSession
	err := row.Scan(&session.LoginKey, &session.Status, &session.UserID, &session.Token, &session.ExpiresAt, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get scan login session: %w", err)
	}
	return &session, nil
}

func (r *repository) UpdateScanLoginSession(ctx context.Context, session *ScanLoginSession) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE scan_login_sessions SET status = $1, user_id = $2, token = $3, updated_at = NOW() WHERE login_key = $4`,
		session.Status, session.UserID, session.Token, session.LoginKey,
	)
	if err != nil {
		return fmt.Errorf("update scan login session: %w", err)
	}
	return nil
}

func (r *repository) CreateWxPayConfig(ctx context.Context, config *WxPayConfig) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	return exec.QueryRowContext(ctx,
		`INSERT INTO wx_pay_config (config_type, app_id, mch_id, api_key, api_v3_key, serial_no, private_key, public_key, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`,
		config.ConfigType, config.AppID, config.MchID, config.APIKey, config.APIV3Key, config.SerialNo, config.PrivateKey, config.PublicKey, config.Status,
	).Scan(&config.ID)
}

func (r *repository) GetWxPayConfig(ctx context.Context, configType string) (*WxPayConfig, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, config_type, app_id, mch_id, api_key, api_v3_key, serial_no, private_key, public_key, status, created_at, updated_at FROM wx_pay_config WHERE config_type = $1 AND status = 1`,
		configType,
	)
	var config WxPayConfig
	err := row.Scan(&config.ID, &config.ConfigType, &config.AppID, &config.MchID, &config.APIKey, &config.APIV3Key, &config.SerialNo, &config.PrivateKey, &config.PublicKey, &config.Status, &config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get wx pay config: %w", err)
	}
	return &config, nil
}

func (r *repository) GetWxPayConfigByID(ctx context.Context, id int64) (*WxPayConfig, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	row := exec.QueryRowContext(ctx,
		`SELECT id, config_type, app_id, mch_id, api_key, api_v3_key, serial_no, private_key, public_key, status, created_at, updated_at FROM wx_pay_config WHERE id = $1`,
		id,
	)
	var config WxPayConfig
	err := row.Scan(&config.ID, &config.ConfigType, &config.AppID, &config.MchID, &config.APIKey, &config.APIV3Key, &config.SerialNo, &config.PrivateKey, &config.PublicKey, &config.Status, &config.CreatedAt, &config.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get wx pay config by id: %w", err)
	}
	return &config, nil
}

func (r *repository) ListWxPayConfigs(ctx context.Context, page, pageSize int, configType *string) ([]*WxPayConfig, int, error) {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	countQuery := "SELECT COUNT(*) FROM wx_pay_config WHERE 1=1"
	args := []interface{}{}
	argCount := 0

	if configType != nil {
		argCount++
		countQuery += fmt.Sprintf(" AND config_type = $%d", argCount)
		args = append(args, *configType)
	}

	var total int
	if err := exec.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count wx pay configs: %w", err)
	}

	dataQuery := `SELECT id, config_type, app_id, mch_id, api_key, api_v3_key, serial_no, private_key, public_key, status, created_at, updated_at FROM wx_pay_config WHERE 1=1`
	dataArgs := []interface{}{}
	argCount = 0

	if configType != nil {
		argCount++
		dataQuery += fmt.Sprintf(" AND config_type = $%d", argCount)
		dataArgs = append(dataArgs, *configType)
	}

	dataQuery += " ORDER BY id DESC"
	argCount++
	dataQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	dataArgs = append(dataArgs, pageSize, (page-1)*pageSize)

	rows, err := exec.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list wx pay configs: %w", err)
	}
	defer rows.Close()

	var configs []*WxPayConfig
	for rows.Next() {
		var c WxPayConfig
		if err := rows.Scan(&c.ID, &c.ConfigType, &c.AppID, &c.MchID, &c.APIKey, &c.APIV3Key, &c.SerialNo, &c.PrivateKey, &c.PublicKey, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan wx pay config: %w", err)
		}
		configs = append(configs, &c)
	}
	return configs, total, nil
}

func (r *repository) UpdateWxPayConfig(ctx context.Context, config *WxPayConfig) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE wx_pay_config SET app_id = $1, mch_id = $2, api_key = $3, api_v3_key = $4, serial_no = $5, private_key = $6, public_key = $7, status = $8, updated_at = NOW() WHERE id = $9`,
		config.AppID, config.MchID, config.APIKey, config.APIV3Key, config.SerialNo, config.PrivateKey, config.PublicKey, config.Status, config.ID,
	)
	if err != nil {
		return fmt.Errorf("update wx pay config: %w", err)
	}
	return nil
}

func (r *repository) UpdateWxPayConfigStatus(ctx context.Context, id int64, status int) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx,
		`UPDATE wx_pay_config SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update wx pay config status: %w", err)
	}
	return nil
}

func (r *repository) DeleteWxPayConfig(ctx context.Context, id int64) error {
	exec := database.ExecutorFromContext(ctx, r.dbtx)
	_, err := exec.ExecContext(ctx, `DELETE FROM wx_pay_config WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete wx pay config: %w", err)
	}
	return nil
}
