-- name: GetUserByID :one
SELECT id, public_id, email, password_hash, status, password_changed_at, last_login_at, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetByEmail :one
SELECT id, public_id, email, password_hash, status, password_changed_at, last_login_at, created_at, updated_at
FROM users
WHERE email = $1;

-- name: CreateBalanceLog :one
INSERT INTO user_balance_logs (user_id, change_type, amount, balance_after, related_id, related_no, description, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- name: GetUserBalanceLogs :many
SELECT id, user_id, change_type, amount, balance_after, related_id, related_no, description, created_at
FROM user_balance_logs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUserBalanceLogs :one
SELECT COUNT(*) FROM user_balance_logs
WHERE user_id = $1;

-- name: GetUserLevel :one
SELECT id, name, min_consumption, discount_rate, benefits, status, created_at, updated_at
FROM user_levels
WHERE id = $1;

-- name: CreateUserLevelLog :one
INSERT INTO user_level_logs (user_id, old_level_id, new_level_id, change_reason, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: CreatePurchaseRecord :one
INSERT INTO user_purchase_records (user_id, goods_id, order_id, quantity, purchase_time)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: GetUserPurchaseCount :one
SELECT COUNT(*) FROM user_purchase_records
WHERE user_id = $1 AND goods_id = $2;
