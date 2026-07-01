-- name: ListCategories :many
SELECT id, parent_id, name, icon, sort, status, created_at, updated_at
FROM goods_categories
ORDER BY sort ASC, id ASC;

-- name: GetCategory :one
SELECT id, parent_id, name, icon, sort, status, created_at, updated_at
FROM goods_categories
WHERE id = $1;

-- name: CreateCategory :one
INSERT INTO goods_categories (parent_id, name, icon, sort, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id;

-- name: CreateGoods :one
INSERT INTO goods (category_id, platform, name, description, cover_image, billing_mode, status, is_visible, commission_type, commission_rate, min_teacher_level, map_select_enabled, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
RETURNING id;

-- name: GetGoods :one
SELECT id, category_id, platform, name, description, cover_image, billing_mode, status, is_visible, commission_type, commission_rate, min_teacher_level, map_select_enabled, version, created_at, updated_at
FROM goods
WHERE id = $1;

-- name: ListGoods :many
SELECT id, category_id, platform, name, description, cover_image, billing_mode, status, is_visible, commission_type, commission_rate, min_teacher_level, map_select_enabled, version, created_at, updated_at
FROM goods
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountGoods :one
SELECT COUNT(*) FROM goods;

-- name: CreateSKU :one
INSERT INTO goods_skus (goods_id, sku_name, sku_snapshot, price, stock, sort, status, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
RETURNING id;

-- name: GetSKUsByGoodsID :many
SELECT id, goods_id, sku_name, sku_snapshot, price, stock, sort, status, version, created_at, updated_at
FROM goods_skus
WHERE goods_id = $1
ORDER BY sort ASC, id ASC;

-- name: DecreaseStock :exec
UPDATE goods_skus
SET stock = stock - $2, updated_at = NOW()
WHERE id = $1 AND stock >= $2;

-- name: IncreaseStock :exec
UPDATE goods_skus
SET stock = stock + $2, updated_at = NOW()
WHERE id = $1;

-- name: CreateStockLog :one
INSERT INTO goods_sku_stock_logs (sku_id, old_stock, new_stock, change_type, order_id, operator_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
RETURNING id;
