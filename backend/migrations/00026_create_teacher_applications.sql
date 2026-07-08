-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS teacher_applications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(100) NOT NULL,
    mobile VARCHAR(20),
    avatar VARCHAR(500),
    platforms JSONB,
    tags JSONB,
    intro TEXT,
    status SMALLINT DEFAULT 0,
    reason TEXT,
    operator_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_teacher_applications_user_id ON teacher_applications(user_id);
CREATE INDEX IF NOT EXISTS idx_teacher_applications_status ON teacher_applications(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS teacher_applications;
-- +goose StatementEnd
