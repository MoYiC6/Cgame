-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS partner_configs (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    partner_type VARCHAR(64) NOT NULL DEFAULT 'agency',
    commission_rate NUMERIC(5,2) DEFAULT 0,
    fixed_fee NUMERIC(10,2) DEFAULT 0,
    description TEXT,
    contact_name VARCHAR(255),
    contact_phone VARCHAR(64),
    contact_email VARCHAR(255),
    status VARCHAR(32) NOT NULL DEFAULT 'enabled',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS teacher_partners (
    id BIGSERIAL PRIMARY KEY,
    teacher_id BIGINT NOT NULL,
    partner_id BIGINT NOT NULL,
    partner_config_id BIGINT,
    cooperation_type VARCHAR(64) DEFAULT 'exclusive',
    commission_rate NUMERIC(5,2) DEFAULT 0,
    start_date DATE,
    end_date DATE,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    remark TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_partner_configs_status ON partner_configs(status);
CREATE INDEX IF NOT EXISTS idx_teacher_partners_teacher_id ON teacher_partners(teacher_id);
CREATE INDEX IF NOT EXISTS idx_teacher_partners_partner_id ON teacher_partners(partner_id);
CREATE INDEX IF NOT EXISTS idx_teacher_partners_status ON teacher_partners(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS teacher_partners;
DROP TABLE IF EXISTS partner_configs;
-- +goose StatementEnd
