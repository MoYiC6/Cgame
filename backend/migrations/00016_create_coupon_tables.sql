-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS coupon (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    type INTEGER NOT NULL,
    discount_type VARCHAR(50) DEFAULT 'FIXED',
    discount_value NUMERIC(10,2) DEFAULT 0.00,
    face_value NUMERIC(10,2),
    max_discount_amount NUMERIC(10,2),
    min_order_amount NUMERIC(10,2),
    total_quantity INTEGER,
    claimed_quantity INTEGER DEFAULT 0,
    per_user_limit INTEGER DEFAULT 1,
    valid_days INTEGER,
    start_time TIMESTAMPTZ,
    end_time TIMESTAMPTZ,
    applicable_scope TEXT,
    distribution_mode INTEGER DEFAULT 2,
    target_level_ids TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    is_permanent BOOLEAN DEFAULT FALSE,
    restricted_goods_ids TEXT,
    restricted_category_ids TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    total_count INTEGER DEFAULT 0,
    used_count INTEGER DEFAULT 0,
    min_amount NUMERIC(10,2),
    max_discount NUMERIC(10,2),
    status SMALLINT DEFAULT 1,
    create_time TIMESTAMPTZ DEFAULT NOW(),
    update_time TIMESTAMPTZ DEFAULT NOW(),
    deleted SMALLINT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS user_coupon (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    coupon_id BIGINT NOT NULL REFERENCES coupon(id) ON DELETE CASCADE,
    status INTEGER DEFAULT 0,
    source VARCHAR(50),
    order_id BIGINT,
    claimed_at TIMESTAMPTZ,
    used_at TIMESTAMPTZ,
    expire_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE coupon ADD COLUMN IF NOT EXISTS name VARCHAR(200);
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS type INTEGER;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS discount_type VARCHAR(50) DEFAULT 'FIXED';
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS discount_value NUMERIC(10,2) DEFAULT 0.00;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS face_value NUMERIC(10,2);
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS max_discount_amount NUMERIC(10,2);
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS min_order_amount NUMERIC(10,2);
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS total_quantity INTEGER;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS claimed_quantity INTEGER DEFAULT 0;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS per_user_limit INTEGER DEFAULT 1;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS valid_days INTEGER;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS start_time TIMESTAMPTZ;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS end_time TIMESTAMPTZ;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS applicable_scope TEXT;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS distribution_mode INTEGER DEFAULT 2;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS target_level_ids TEXT;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS is_permanent BOOLEAN DEFAULT FALSE;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS restricted_goods_ids TEXT;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS restricted_category_ids TEXT;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS total_count INTEGER DEFAULT 0;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS used_count INTEGER DEFAULT 0;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS min_amount NUMERIC(10,2);
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS max_discount NUMERIC(10,2);
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS status SMALLINT DEFAULT 1;
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS create_time TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS update_time TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE coupon ADD COLUMN IF NOT EXISTS deleted SMALLINT DEFAULT 0;

ALTER TABLE user_coupon ADD COLUMN IF NOT EXISTS user_id BIGINT;
ALTER TABLE user_coupon ADD COLUMN IF NOT EXISTS coupon_id BIGINT;
ALTER TABLE user_coupon ADD COLUMN IF NOT EXISTS status INTEGER DEFAULT 0;
ALTER TABLE user_coupon ADD COLUMN IF NOT EXISTS source VARCHAR(50);
ALTER TABLE user_coupon ADD COLUMN IF NOT EXISTS order_id BIGINT;
ALTER TABLE user_coupon ADD COLUMN IF NOT EXISTS claimed_at TIMESTAMPTZ;
ALTER TABLE user_coupon ADD COLUMN IF NOT EXISTS used_at TIMESTAMPTZ;
ALTER TABLE user_coupon ADD COLUMN IF NOT EXISTS expire_at TIMESTAMPTZ;
ALTER TABLE user_coupon ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_coupon_enabled ON coupon(enabled);
CREATE INDEX IF NOT EXISTS idx_coupon_distribution_mode ON coupon(distribution_mode);
CREATE INDEX IF NOT EXISTS idx_coupon_is_permanent ON coupon(is_permanent);
CREATE INDEX IF NOT EXISTS idx_coupon_start_end_time ON coupon(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_user_coupon_user_id ON user_coupon(user_id);
CREATE INDEX IF NOT EXISTS idx_user_coupon_coupon_id ON user_coupon(coupon_id);
CREATE INDEX IF NOT EXISTS idx_user_coupon_status ON user_coupon(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_coupon;
DROP TABLE IF EXISTS coupon;
-- +goose StatementEnd
