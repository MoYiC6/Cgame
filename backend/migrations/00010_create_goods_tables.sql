-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS goods_categories (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT DEFAULT 0,
    name VARCHAR(100) NOT NULL,
    icon VARCHAR(255),
    sort INTEGER DEFAULT 0,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS goods (
    id BIGSERIAL PRIMARY KEY,
    category_id BIGINT,
    platform VARCHAR(100),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    cover_image VARCHAR(500),
    billing_mode VARCHAR(50),
    status SMALLINT DEFAULT 1,
    is_visible BOOLEAN DEFAULT TRUE,
    commission_type VARCHAR(50),
    commission_rate NUMERIC(5,2),
    min_teacher_level INTEGER DEFAULT 0,
    map_select_enabled BOOLEAN DEFAULT FALSE,
    version INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS goods_skus (
    id BIGSERIAL PRIMARY KEY,
    goods_id BIGINT NOT NULL,
    sku_name VARCHAR(255) NOT NULL,
    sku_snapshot JSONB,
    price NUMERIC(10,2) NOT NULL,
    stock INTEGER DEFAULT 0,
    sort INTEGER DEFAULT 0,
    status SMALLINT DEFAULT 1,
    version INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS goods_specs (
    id BIGSERIAL PRIMARY KEY,
    goods_id BIGINT NOT NULL,
    spec_name VARCHAR(100) NOT NULL,
    spec_values JSONB,
    sort INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS goods_sku_stock_logs (
    id BIGSERIAL PRIMARY KEY,
    sku_id BIGINT NOT NULL,
    old_stock INTEGER,
    new_stock INTEGER,
    change_type VARCHAR(50),
    order_id BIGINT,
    operator_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS purchase_limit_rules (
    id BIGSERIAL PRIMARY KEY,
    goods_id BIGINT,
    limit_type VARCHAR(50) DEFAULT 'per_user',
    limit_count INTEGER DEFAULT 1,
    limit_period INTEGER DEFAULT 86400,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    status SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_goods_category_id ON goods(category_id);
CREATE INDEX IF NOT EXISTS idx_goods_status ON goods(status);
CREATE INDEX IF NOT EXISTS idx_goods_skus_goods_id ON goods_skus(goods_id);
CREATE INDEX IF NOT EXISTS idx_goods_specs_goods_id ON goods_specs(goods_id);
CREATE INDEX IF NOT EXISTS idx_goods_sku_stock_logs_sku_id ON goods_sku_stock_logs(sku_id);
CREATE INDEX IF NOT EXISTS idx_purchase_limit_rules_goods_id ON purchase_limit_rules(goods_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS purchase_limit_rules;
DROP TABLE IF EXISTS goods_sku_stock_logs;
DROP TABLE IF EXISTS goods_specs;
DROP TABLE IF EXISTS goods_skus;
DROP TABLE IF EXISTS goods;
DROP TABLE IF EXISTS goods_categories;
-- +goose StatementEnd
