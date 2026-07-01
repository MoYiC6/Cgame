-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS file_categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description VARCHAR(500),
    sort INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS files (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,
    category_id BIGINT,
    display_name VARCHAR(255),
    original_name VARCHAR(255),
    url VARCHAR(500) NOT NULL,
    file_id VARCHAR(255),
    file_hash VARCHAR(64),
    type VARCHAR(100),
    size BIGINT,
    provider VARCHAR(50) DEFAULT 'qiniu',
    status SMALLINT DEFAULT 1,
    description TEXT,
    sort INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted SMALLINT DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
CREATE INDEX IF NOT EXISTS idx_files_category_id ON files(category_id);
CREATE INDEX IF NOT EXISTS idx_files_file_hash ON files(file_hash);
CREATE INDEX IF NOT EXISTS idx_files_status ON files(status);
CREATE INDEX IF NOT EXISTS idx_files_type ON files(type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS file_categories;
-- +goose StatementEnd
