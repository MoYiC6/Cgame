-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_oauth (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    platform VARCHAR(50) NOT NULL,
    open_id VARCHAR(128) NOT NULL,
    union_id VARCHAR(128),
    nickname VARCHAR(100),
    avatar VARCHAR(500),
    session_key TEXT,
    phone VARCHAR(20),
    bound_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, platform),
    UNIQUE(platform, open_id)
);

CREATE TABLE IF NOT EXISTS user_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS scan_login_sessions (
    login_key VARCHAR(64) PRIMARY KEY,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    user_id BIGINT,
    token TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS wx_pay_config (
    id BIGSERIAL PRIMARY KEY,
    config_type VARCHAR(50) NOT NULL,
    app_id VARCHAR(64) NOT NULL,
    mch_id VARCHAR(64) NOT NULL,
    api_key TEXT,
    api_v3_key TEXT,
    serial_no VARCHAR(64),
    private_key TEXT,
    public_key TEXT,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_oauth_user_id ON user_oauth(user_id);
CREATE INDEX IF NOT EXISTS idx_user_oauth_platform_openid ON user_oauth(platform, open_id);
CREATE INDEX IF NOT EXISTS idx_user_tokens_user_id ON user_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_user_tokens_access_token ON user_tokens(access_token);
CREATE INDEX IF NOT EXISTS idx_wx_pay_config_type ON wx_pay_config(config_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wx_pay_config;
DROP TABLE IF EXISTS scan_login_sessions;
DROP TABLE IF EXISTS user_tokens;
DROP TABLE IF EXISTS user_oauth;
-- +goose StatementEnd
