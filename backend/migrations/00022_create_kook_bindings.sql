-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS kook_bindings (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    kook_user_id VARCHAR(255),
    kook_nickname VARCHAR(255),
    bind_code VARCHAR(32) NOT NULL,
    bound_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kook_bindings_user_id ON kook_bindings(user_id);
CREATE INDEX IF NOT EXISTS idx_kook_bindings_bind_code ON kook_bindings(bind_code);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS kook_bindings;
-- +goose StatementEnd
