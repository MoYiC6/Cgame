-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS system_settings (
    id BIGSERIAL PRIMARY KEY,
    setting_key VARCHAR(100) UNIQUE NOT NULL,
    setting_value TEXT,
    setting_type VARCHAR(50) DEFAULT 'string',
    category VARCHAR(50),
    description VARCHAR(500),
    is_public SMALLINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS partner_config (
    id BIGSERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT,
    description TEXT,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS faceid_config (
    id BIGSERIAL PRIMARY KEY,
    secret_id VARCHAR(255) NOT NULL,
    secret_key VARCHAR(255) NOT NULL,
    rule_id VARCHAR(255) NOT NULL,
    region VARCHAR(50) DEFAULT 'ap-guangzhou',
    redirect_url VARCHAR(500),
    is_enabled SMALLINT DEFAULT 1,
    remark TEXT,
    manual_enabled SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted SMALLINT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS realname_verify_log (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    operator_id BIGINT,
    operator_type VARCHAR(20),
    detail TEXT,
    ip_address VARCHAR(50),
    submitted_name VARCHAR(100),
    submitted_id_card VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted SMALLINT DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_system_settings_key ON system_settings(setting_key);
CREATE INDEX IF NOT EXISTS idx_partner_config_key ON partner_config(config_key);
CREATE INDEX IF NOT EXISTS idx_faceid_config_enabled ON faceid_config(is_enabled, deleted);
CREATE INDEX IF NOT EXISTS idx_realname_verify_log_user ON realname_verify_log(user_id, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS realname_verify_log;
DROP TABLE IF EXISTS faceid_config;
DROP TABLE IF EXISTS partner_config;
DROP TABLE IF EXISTS system_settings;
-- +goose StatementEnd
