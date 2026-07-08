-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS invite_records (
    id BIGSERIAL PRIMARY KEY,
    inviter_id BIGINT NOT NULL,
    invitee_id BIGINT,
    invite_code VARCHAR(32),
    inviter_reward_coupon_id BIGINT,
    invitee_reward_coupon_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS teacher_invite_codes (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(32) NOT NULL UNIQUE,
    status VARCHAR(32) NOT NULL DEFAULT 'unused',
    remark VARCHAR(255),
    expire_time TIMESTAMPTZ,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    used_by BIGINT,
    used_time TIMESTAMPTZ,
    teacher_id BIGINT,
    revoked_by BIGINT,
    revoked_time TIMESTAMPTZ,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_invite_records_inviter_id ON invite_records(inviter_id);
CREATE INDEX IF NOT EXISTS idx_invite_records_invitee_id ON invite_records(invitee_id);
CREATE INDEX IF NOT EXISTS idx_invite_records_invite_code ON invite_records(invite_code);
CREATE INDEX IF NOT EXISTS idx_teacher_invite_codes_code ON teacher_invite_codes(code);
CREATE INDEX IF NOT EXISTS idx_teacher_invite_codes_status ON teacher_invite_codes(status);
CREATE INDEX IF NOT EXISTS idx_teacher_invite_codes_created_by ON teacher_invite_codes(created_by);
CREATE INDEX IF NOT EXISTS idx_teacher_invite_codes_used_by ON teacher_invite_codes(used_by);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS teacher_invite_codes;
DROP TABLE IF EXISTS invite_records;
-- +goose StatementEnd
