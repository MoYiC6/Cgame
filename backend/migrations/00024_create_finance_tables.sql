-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS operator_commissions (
    id BIGSERIAL PRIMARY KEY,
    operator_id BIGINT NOT NULL,
    order_id BIGINT,
    amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    balance NUMERIC(10,2) NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS operator_withdrawals (
    id BIGSERIAL PRIMARY KEY,
    operator_id BIGINT NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    admin_remark TEXT,
    processed_by BIGINT,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS balance_details (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    change_type VARCHAR(64) NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    balance NUMERIC(10,2) NOT NULL,
    remark TEXT,
    related_id BIGINT,
    related_type VARCHAR(64),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_operator_commissions_operator_id ON operator_commissions(operator_id);
CREATE INDEX IF NOT EXISTS idx_operator_withdrawals_operator_id ON operator_withdrawals(operator_id);
CREATE INDEX IF NOT EXISTS idx_balance_details_user_id ON balance_details(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS balance_details;
DROP TABLE IF EXISTS operator_withdrawals;
DROP TABLE IF EXISTS operator_commissions;
-- +goose StatementEnd
