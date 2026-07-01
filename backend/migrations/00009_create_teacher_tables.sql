-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS teachers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(100),
    mobile VARCHAR(20),
    avatar VARCHAR(500),
    status SMALLINT DEFAULT 1,
    rating NUMERIC(3,2) DEFAULT 5.0,
    order_count INTEGER DEFAULT 0,
    deposit NUMERIC(10,2) DEFAULT 0,
    balance NUMERIC(10,2) DEFAULT 0,
    platforms JSONB,
    tags JSONB,
    goods_ids JSONB,
    auto_status_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS teacher_levels (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    min_orders INTEGER DEFAULT 0,
    commission_rate NUMERIC(5,2) DEFAULT 1.0,
    priority INTEGER DEFAULT 0,
    status SMALLINT DEFAULT 1,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS teacher_level_goods (
    id BIGSERIAL PRIMARY KEY,
    level_id BIGINT NOT NULL,
    goods_id BIGINT NOT NULL,
    commission_rate NUMERIC(5,2) DEFAULT 1.0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS teacher_status_logs (
    id BIGSERIAL PRIMARY KEY,
    teacher_id BIGINT NOT NULL,
    old_status SMALLINT,
    new_status SMALLINT,
    reason TEXT,
    operator_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS teacher_balance_logs (
    id BIGSERIAL PRIMARY KEY,
    teacher_id BIGINT NOT NULL,
    change_type VARCHAR(50) NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    balance_after NUMERIC(10,2) NOT NULL,
    related_id BIGINT,
    related_no VARCHAR(100),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_teachers_user_id ON teachers(user_id);
CREATE INDEX IF NOT EXISTS idx_teachers_status ON teachers(status);
CREATE INDEX IF NOT EXISTS idx_teacher_status_logs_teacher_id ON teacher_status_logs(teacher_id);
CREATE INDEX IF NOT EXISTS idx_teacher_balance_logs_teacher_id ON teacher_balance_logs(teacher_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS teacher_balance_logs;
DROP TABLE IF EXISTS teacher_status_logs;
DROP TABLE IF EXISTS teacher_level_goods;
DROP TABLE IF EXISTS teacher_levels;
DROP TABLE IF EXISTS teachers;
-- +goose StatementEnd
