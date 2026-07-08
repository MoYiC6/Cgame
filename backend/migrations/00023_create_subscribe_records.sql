-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS subscribe_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    template_id VARCHAR(255) NOT NULL,
    subscribed_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, template_id)
);

CREATE INDEX IF NOT EXISTS idx_subscribe_records_user_id ON subscribe_records(user_id);
CREATE INDEX IF NOT EXISTS idx_subscribe_records_template_id ON subscribe_records(template_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS subscribe_records;
-- +goose StatementEnd
