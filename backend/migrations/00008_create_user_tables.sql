-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_balance_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    change_type VARCHAR(50) NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    balance_after NUMERIC(10,2) NOT NULL,
    related_id BIGINT,
    related_no VARCHAR(100),
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_levels (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    min_consumption NUMERIC(10,2) DEFAULT 0,
    discount_rate NUMERIC(5,2) DEFAULT 1.0,
    benefits JSONB,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_level_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    old_level_id BIGINT,
    new_level_id BIGINT,
    change_reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_purchase_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    goods_id BIGINT,
    order_id BIGINT,
    quantity INTEGER DEFAULT 1,
    purchase_time TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_balance_logs_user_id ON user_balance_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_user_level_logs_user_id ON user_level_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_user_purchase_records_user_id ON user_purchase_records(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_purchase_records;
DROP TABLE IF EXISTS user_level_logs;
DROP TABLE IF EXISTS user_levels;
DROP TABLE IF EXISTS user_balance_logs;
-- +goose StatementEnd
