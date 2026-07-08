-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS recharge_records (
    id BIGSERIAL PRIMARY KEY,
    recharge_no VARCHAR(64) NOT NULL UNIQUE,
    user_id BIGINT NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    gift_amount NUMERIC(10,2) DEFAULT 0,
    total_amount NUMERIC(10,2) NOT NULL,
    pay_amount NUMERIC(10,2) DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    pay_channel VARCHAR(32),
    pay_time TIMESTAMPTZ,
    callback_time TIMESTAMPTZ,
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS recharge_rebate_rules (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    min_amount NUMERIC(10,2) NOT NULL,
    gift_rate NUMERIC(5,2) DEFAULT 0,
    gift_amount NUMERIC(10,2) DEFAULT 0,
    enabled BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_recharge_records_user_id ON recharge_records(user_id);
CREATE INDEX IF NOT EXISTS idx_recharge_records_status ON recharge_records(status);
CREATE INDEX IF NOT EXISTS idx_recharge_records_recharge_no ON recharge_records(recharge_no);
CREATE INDEX IF NOT EXISTS idx_recharge_records_created_at ON recharge_records(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recharge_rebate_rules;
DROP TABLE IF EXISTS recharge_records;
-- +goose StatementEnd
