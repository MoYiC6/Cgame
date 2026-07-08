-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sys_user (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(64) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE,
    password VARCHAR(255),
    nickname VARCHAR(128),
    real_name VARCHAR(128),
    mobile VARCHAR(20) UNIQUE,
    email_verified SMALLINT DEFAULT 0,
    avatar VARCHAR(500),
    gender SMALLINT DEFAULT 0,
    birthday DATE,
    intro TEXT,
    province VARCHAR(64),
    city VARCHAR(64),
    district VARCHAR(64),
    wechat VARCHAR(64),
    wechat_union_id VARCHAR(64),
    wechat_mp_open_id VARCHAR(64),
    wechat_h5_open_id VARCHAR(64),
    alipay_open_id VARCHAR(64),
    status SMALLINT DEFAULT 1,
    is_teacher SMALLINT DEFAULT 0,
    id_card VARCHAR(18),
    id_card_front VARCHAR(500),
    id_card_back VARCHAR(500),
    real_name_status SMALLINT DEFAULT 0,
    real_name_verify_type VARCHAR(32),
    real_name_submit_time TIMESTAMPTZ,
    real_name_verify_time TIMESTAMPTZ,
    real_name_reject_reason VARCHAR(255),
    login_failed_count INT DEFAULT 0,
    last_login_time TIMESTAMPTZ,
    last_login_ip VARCHAR(64),
    last_login_platform VARCHAR(32),
    password_updated_at TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ,
    register_ip VARCHAR(64),
    register_platform VARCHAR(32),
    register_source VARCHAR(32),
    balance NUMERIC(12,2) DEFAULT 0,
    frozen_balance NUMERIC(12,2) DEFAULT 0,
    total_recharge NUMERIC(12,2) DEFAULT 0,
    total_consumption NUMERIC(12,2) DEFAULT 0,
    level_id BIGINT,
    deleted SMALLINT DEFAULT 0,
    create_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    update_time TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sys_user_status ON sys_user(status);
CREATE INDEX IF NOT EXISTS idx_sys_user_is_teacher ON sys_user(is_teacher);
CREATE INDEX IF NOT EXISTS idx_sys_user_level_id ON sys_user(level_id);
CREATE INDEX IF NOT EXISTS idx_sys_user_mobile ON sys_user(mobile);
CREATE INDEX IF NOT EXISTS idx_sys_user_create_time ON sys_user(create_time);
CREATE INDEX IF NOT EXISTS idx_sys_user_deleted ON sys_user(deleted);

-- user_login_logs 表
CREATE TABLE IF NOT EXISTS user_login_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    login_type VARCHAR(32),
    ip_address VARCHAR(64),
    user_agent VARCHAR(500),
    login_status VARCHAR(32),
    fail_reason VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_login_logs_user_id ON user_login_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_user_login_logs_created_at ON user_login_logs(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_login_logs;
DROP TABLE IF EXISTS sys_user;
-- +goose StatementEnd
