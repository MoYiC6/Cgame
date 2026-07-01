-- name: CreateOrder :one
INSERT INTO orders (order_no, user_id, status, total_amount, pay_amount, discount_amount, goods_id, sku_name, quantity, remark, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
RETURNING id;

-- name: GetOrderByID :one
SELECT id, order_no, user_id, status, total_amount, pay_amount, discount_amount, goods_id, sku_name, quantity, remark, pay_at, completed_at, cancelled_at, created_at, updated_at
FROM orders
WHERE id = $1;

-- name: GetOrderByNo :one
SELECT id, order_no, user_id, status, total_amount, pay_amount, discount_amount, goods_id, sku_name, quantity, remark, pay_at, completed_at, cancelled_at, created_at, updated_at
FROM orders
WHERE order_no = $1;

-- name: ListOrders :many
SELECT id, order_no, user_id, status, total_amount, pay_amount, discount_amount, goods_id, sku_name, quantity, remark, pay_at, completed_at, cancelled_at, created_at, updated_at
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountOrders :one
SELECT COUNT(*) FROM orders
WHERE user_id = $1;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, goods_id, sku_name, price, quantity, subtotal, created_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
RETURNING id;

-- name: GetOrderItems :many
SELECT id, order_id, goods_id, sku_name, price, quantity, subtotal, created_at
FROM order_items
WHERE order_id = $1
ORDER BY created_at ASC;
